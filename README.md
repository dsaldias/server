# DataAuth CLI

Herramienta CLI para generar automáticamente la configuración inicial de un servidor con integración de autenticación y GraphQL.

Inspirado en el flujo de herramientas como gqlgen, pero enfocado en simplificar la inicialización de proyectos.

---

## 🚀 Instalación

```bash
go get -tool github.com/99designs/gqlgen
go tool gqlgen init
```

```bash
go get -tool github.com/dsaldias/server/generar
```

---

## ⚙️ Uso

Inicializar un nuevo archivo de servidor:

```bash
go tool generar init
```

Esto generará automáticamente un archivo:

```
serverx.go
```

---

## 📦 Qué hace

El comando `init`:

* Detecta automáticamente el módulo del proyecto desde `go.mod`
* Genera un archivo `serverx.go` listo para usar
* Integra:

  * Configuración básica de servidor HTTP
  * Setup de GraphQL
  * Hooks para autenticación (`dataauth`)
  * Conexión a base de datos (según utilidades del proyecto)

---

## 📁 Ejemplo de flujo

```bash
go get -tool github.com/dsaldias/server/generar
go tool generar init
go run serverx.go
```

---

## 🧠 Requisitos

* Proyecto Go con `go.mod` inicializado
* Estructura compatible con GraphQL (por ejemplo carpeta `graph/`)
* Dependencias necesarias instaladas (`gqlgen`, `dataauth`, etc.)

---

## ⚠️ Notas

* Si `serverx.go` ya existe, se recomienda eliminarlo o renombrarlo antes de ejecutar `init`
* El archivo generado puede ser modificado manualmente según necesidades del proyecto
* Asegúrate de tener configuradas correctamente las utilidades como conexión a base de datos

---

## 🛠️ Roadmap (ideas futuras)

* Soporte para flags (`--port`, `--name`, etc.)
* Uso de templates para generación más flexible
* Generación de estructura completa de proyecto
* Validaciones automáticas de entorno

---

## 🤝 Contribuciones

Las contribuciones son bienvenidas. Si tenés ideas o mejoras, abrí un issue o pull request.

---

## 📄 Licencia

GNU GENERAL PUBLIC LICENSE
