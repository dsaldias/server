package xlogin

import (
	"database/sql"
	"fmt"
)

func GuardarDatosPortal(db *sql.DB, u, p string, res *Response) error {
	fmt.Println("guardando")
	sq := `
	INSERT INTO rbac_usuarios(id,nombres, apellido1, apellido2, username, password)
	VALUES (?, ?, ?, ?, ?, SHA2(?, 256))
	on duplicate key update nombres=values(nombres), apellido1=values(apellido1), apellido2=values(apellido2), password = values(password)
	`

	_, err := db.Exec(
		sq,
		res.Data.Me.DatosPersonales.ID,
		res.Data.Me.DatosPersonales.Nombres,
		res.Data.Me.DatosPersonales.PriApellido,
		res.Data.Me.DatosPersonales.SegApellido,
		u,
		p,
	)

	if err != nil {
		return err
	}

	for _, c := range res.Data.Me.Cargos {
		sq = `insert into rbac_unidades(id, nombre) values(?,?) on duplicate key update nombre=values(nombre)`
		_, err := db.Exec(sq, c.Sede.ID, c.Sede.Nombre)
		if err != nil {
			return err
		}

		sq = `insert into rbac_roles(id,nombre) values (?,?) on duplicate key update nombre=values(nombre)`
		_, err = db.Exec(sq, c.Cargo.ID, c.Cargo.Nombre)
		if err != nil {
			return err
		}

		sq = `insert ignore into rbac_rol_usuario_unidades(rol_id,usuario_id,unidad_id) values (?,?,?) `
		_, err = db.Exec(sq, c.Cargo.ID, res.Data.Me.DatosPersonales.ID, c.Sede.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
