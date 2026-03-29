package usuarios

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/dsaldias/server/graph_auth/model"
)

func parseRow(row *sql.Row, t *model.Usuario) error {
	return row.Scan(
		&t.ID,
		&t.Nombres,
		&t.Apellido1,
		&t.Apellido2,
		&t.Documento,
		&t.Celular,
		&t.Correo,
		&t.Sexo,
		&t.Direccion,
		&t.Estado,
		&t.Username,
		&t.LastLogin,
		&t.OauthID,
		&t.FotoURL,
		&t.Latitud,
		&t.Longitud,
		&t.FechaRegistro,
		&t.FechaUpdate,
	)
}
func parseRowRU(row *sql.Row, t *model.ResponseUsuario) error {
	return row.Scan(
		&t.ID,
		&t.Nombres,
		&t.Apellido1,
		&t.Apellido2,
		&t.Documento,
		&t.Celular,
		&t.Correo,
		&t.Sexo,
		&t.Direccion,
		&t.Estado,
		&t.Username,
		&t.LastLogin,
		&t.OauthID,
		&t.FotoURL,
		&t.Latitud,
		&t.Longitud,
		&t.FechaRegistro,
		&t.FechaUpdate,
	)
}

func parseRows(rows *sql.Rows, t *model.Usuario) error {
	return rows.Scan(
		&t.ID,
		&t.Nombres,
		&t.Apellido1,
		&t.Apellido2,
		&t.Documento,
		&t.Celular,
		&t.Correo,
		&t.Sexo,
		&t.Direccion,
		&t.Estado,
		&t.Username,
		&t.LastLogin,
		&t.OauthID,
		&t.FotoURL,
		&t.Latitud,
		&t.Longitud,
		&t.FechaRegistro,
		&t.FechaUpdate,
	)
}

func validar_campos(input model.NewUsuario) error {
	if len(input.Nombres) > 60 {
		return errors.New("nombres excede 60 caracteres")
	}
	if len(input.Apellido1) > 30 {
		return errors.New("apellido1 excede 30 caracteres")
	}
	if input.Apellido2 != nil && len(*input.Apellido2) > 30 {
		return errors.New("apellido2 excede 30 caracteres")
	}
	if input.Documento != nil && len(*input.Documento) > 30 {
		return errors.New("documento excede 30 caracteres")
	}
	if input.Celular != nil && len(*input.Celular) > 20 {
		return errors.New("celular excede 20 caracteres")
	}
	if input.Correo != nil && len(*input.Correo) > 100 {
		return errors.New("correo excede 100 caracteres")
	}
	if input.Sexo != nil && (*input.Sexo != "M" && *input.Sexo != "F") {
		return errors.New("sexo debe ser M | F")
	}
	if input.Direccion != nil && len(*input.Direccion) > 100 {
		return errors.New("direccion excede 100 caracteres")
	}
	if len(input.Username) > 30 {
		return errors.New("username excede 30 caracteres")
	}
	if len(input.Password) > 30 {
		return errors.New("password excede 30 caracteres")
	}

	return nil
}

func esCaracterValido(r rune) error {
	if unicode.IsSpace(r) {
		return errors.New("contiene espacios")
	}

	if r == 'ñ' || r == 'Ñ' {
		return errors.New("contiene ñ")
	}

	tildes := "áéíóúÁÉÍÓÚ"
	for _, t := range tildes {
		if r == t {
			return errors.New("contiene tildes")
		}
	}

	return nil
}

func validarCadena(cadena, field string) error {
	for _, r := range cadena {
		if err := esCaracterValido(r); err != nil {
			return errors.New(field + ": " + err.Error())
		}
	}
	return nil
}

func permisos_obligatorios(rbac_roles []*model.RolUnidad, permisosueltos []string) error {
	if len(rbac_roles) == 0 && len(permisosueltos) == 0 {
		return errors.New("selecciona al menos un rol o un permiso")
	}
	return nil
}

func oauth_emails_permitidos(email *string) error {
	emails := os.Getenv("OAUTH_EMAILS_PERM")
	emails = strings.Trim(emails, " ")
	perms := strings.Split(emails, ",")

	if len(emails) == 0 {
		return nil
	}

	if email == nil {
		return errors.New("el correo no debe ser vacio")
	}

	parts := strings.Split(*email, "@")
	if len(parts) != 2 {
		return errors.New("email no válido")
	}
	domain := parts[1]

	for _, perm := range perms {
		if strings.TrimSpace(perm) == domain {
			return nil
		}
	}

	return errors.New("utilice su correo de estos dominios: " + emails)
}

func splitName(name string) (string, string) {
	words := strings.Split(name, " ")
	if len(words) <= 2 {
		return name, ""
	}

	firstPart := strings.Join(words[:2], " ")
	secondPart := strings.Join(words[2:], " ")
	return firstPart, secondPart
}

func cut_string(name string, max int) string {
	if len(name) > 30 {
		return name[:max]
	}
	return name
}

func Ubicacion(lat, lon *float64) (*string, error) {
	if (lat != nil && lon == nil) || (lat == nil && lon != nil) {
		return nil, errors.New("si manda lat o lon, ambos deben tener un valor o ninguno")
	}

	if (lat != nil && lon != nil) || (lat == nil && lon == nil) {
		// Ambos valores están presentes o ambos son nulos.
		if lat != nil && lon != nil {
			point := fmt.Sprintf("POINT(%v %v)", *lat, *lon)
			return &point, nil
		}
	}
	return nil, nil
}
func asignarRoles(tx *sql.Tx, rbac_roles []*model.RolUnidad, userid int64) error {
	user_rols := "replace into `rbac_rol_usuario_unidades`(`rol_id`,`unidad_id`,`usuario_id`) values %s"
	places := make([]string, len(rbac_roles))
	args := make([]interface{}, len(rbac_roles)*3)

	for i, r := range rbac_roles {
		places[i] = "(?,?,?)"
		args[i*3] = r.RolID
		args[i*3+1] = r.UnidadID
		args[i*3+2] = userid
	}

	user_rols = fmt.Sprintf(user_rols, strings.Join(places, ", "))
	_, err := tx.Exec(user_rols, args...)
	return err
}

func asignarPermisos(tx *sql.Tx, permisosSueltos []string, userid int64) error {
	user_perms := "replace into `rbac_usuario_permiso`(`usuario_id`,`metodo`) values %s"
	places2 := make([]string, len(permisosSueltos))
	args2 := make([]interface{}, len(permisosSueltos)*2)

	for i, p := range permisosSueltos {
		places2[i] = "(?,?)"
		args2[i*2] = userid
		args2[i*2+1] = p
	}

	user_perms = fmt.Sprintf(user_perms, strings.Join(places2, ", "))
	_, err := tx.Exec(user_perms, args2...)
	return err
}

func asignarMenus(tx *sql.Tx, menus_sueltos []string, userid int64) error {
	user_perms := "replace into `rbac_menus_usuario`(`usuario_id`,`menu_id`) values %s"
	places2 := make([]string, len(menus_sueltos))
	args2 := make([]any, len(menus_sueltos)*2)

	for i, m := range menus_sueltos {
		places2[i] = "(?,?)"
		args2[i*2] = userid
		args2[i*2+1] = m
	}

	user_perms = fmt.Sprintf(user_perms, strings.Join(places2, ", "))
	_, err := tx.Exec(user_perms, args2...)
	return err
}
