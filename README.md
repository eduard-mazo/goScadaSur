# Survalent DB Launcher & Executor

## 📝 Resumen

Este proyecto consiste en dos componentes principales que trabajan juntos para interactuar con una base de datos de Survalent:

1.  **Launcher (Go):** Una aplicación de línea de comandos (CLI) robusta construida con `Cobra`. Su función es actuar como una interfaz amigable para el usuario, capturar las solicitudes y ejecutar el programa de C#.
2.  **Executor (C#):** Una aplicación de consola que recibe comandos desde el *launcher*, se conecta a la base de datos de Survalent utilizando la librería propietaria `ZStcConn.dll`, ejecuta las consultas y devuelve los resultados en un formato JSON seguro.

El sistema garantiza la **integridad de los datos** mediante un checksum SHA256 y exporta los resultados a un archivo **CSV** para su fácil análisis.

## ✨ Características Principales

* **Interfaz de Comandos Clara:** Usa subcomandos (`station-search`, `direct-query`) para diferenciar las operaciones.
* **Gestión de Credenciales Flexible:** Permite pasar credenciales mediante flags (`-u`, `-p`) para scripting o solicitarlas de forma interactiva y segura.
* **Verificación de Integridad:** El *executor* de C# calcula un checksum SHA256 del *payload* de datos. El *launcher* de Go verifica este checksum antes de procesar los datos, previniendo la corrupción.
* **Exportación a CSV:** Los resultados de cualquier consulta exitosa se guardan automáticamente en un archivo `.csv` con marca de tiempo.
* **Arquitectura Desacoplada:** El *launcher* y el *executor* se comunican a través de flujos estándar (stdin, stdout, stderr), lo que los mantiene modulares e independientes.

## 🛠️ Prerrequisitos

Para compilar y ejecutar este proyecto, necesitarás:
* SDK de .NET (v6.0 o superior)
* Compilador de Go (v1.18 o superior)
* La librería `ZStcConn.dll` (y sus dependencias) en el directorio de salida del proyecto de C#.
* Acceso de red al servidor de la base de datos de Survalent.

## 🚀 Cómo Compilar

Sigue estos pasos desde la raíz del proyecto.

1.  **Compilar el Executor de C#:**
    ```shell
    # Navega al directorio del proyecto de C#
    cd ruta/a/tu/proyecto_cs
    
    # Compila el proyecto
    dotnet build -c Release
    
    # Asegúrate de que survalentDB.exe y ZStcConn.dll estén en el mismo directorio de salida
    # (p. ej., ./bin/Release/net6.0/)
    ```

2.  **Compilar el Launcher de Go:**
    ```shell
    # Navega al directorio del proyecto de Go
    cd ruta/a/tu/proyecto_go
    
    # Compila el ejecutable
    go build -o launcher.exe .
    ```

3.  **Preparación Final:**
    Copia el ejecutable `launcher.exe` al mismo directorio donde se encuentra `survalentDB.exe`.

## 💻 Uso

El ejecutable `launcher.exe` es la única interfaz que necesitas usar. Tiene dos subcomandos principales.

### 1. `station-search`

Busca una estación por su nombre y recupera una lista combinada de todas sus señales analógicas y digitales, junto con su dirección IEC104.

**Sintaxis:**
`launcher.exe station-search [nombre-de-estacion]`

**Ejemplo:**
```shell
# Buscará estaciones cuyo nombre contenga "MiEstacion"
./launcher.exe station-search "*MiEstacion*" -u mi_usuario -p mi_contraseña
```

### 2. `direct-query`

Ejecuta una consulta SQL directamente en la base de datos. La consulta debe ir entre comillas si contiene espacios.

**Sintaxis:**
`launcher.exe direct-query "[consulta-sql]"`

**Ejemplo:**
```shell
# Seleccionar los primeros 10 puntos analógicos
./launcher.exe direct-query "SELECT TOP 10 Pkey, Name, Desc FROM AnalogPoints" -u mi_usuario
```
*(Si no se proporciona la contraseña con `-p`, el programa la solicitará de forma segura).*

### Flags Globales

* `-u`, `--user`: Especifica el nombre de usuario para la conexión a la base de datos.
* `-p`, `--password`: Especifica la contraseña. (**Nota:** Es más seguro omitirlo para que el programa lo pida interactivamente).

## 🏛️ Arquitectura y Flujo de Datos

1.  El usuario ejecuta `launcher.exe` con un subcomando y argumentos.
2.  El *launcher* (Go) inicia el proceso `survalentDB.exe` (C#) en segundo plano.
3.  El *launcher* escribe 4 líneas en el `stdin` del proceso hijo:
    * **Modo:** `station_search` o `direct_query`.
    * **Usuario:** El nombre de usuario.
    * **Contraseña:** La contraseña.
    * **Query:** El término de búsqueda o la consulta SQL.
4.  El *executor* (C#) lee esta información, se conecta a la base de datos y ejecuta la lógica correspondiente.
5.  El *executor* serializa el resultado en un `DataTable`, lo convierte a un JSON interno (`payload`), calcula su checksum SHA256 y envuelve ambos en un JSON externo.
6.  Este JSON externo se imprime en el `stdout` del *executor*.
7.  El *launcher* (Go) captura el `stdout`, valida que sea un JSON, y verifica que el checksum recibido coincida con el checksum calculado del `payload`.
8.  Si la integridad es correcta, el *launcher* parsea el `payload` y lo guarda en un archivo `YYYYMMDD_HHMMSS_output.csv`.
9.  Cualquier error del *executor* se envía a `stderr` y es mostrado por el *launcher*.

## 📁 Estructura del Proyecto

```
/
├── launcher/          # Proyecto de Go
│   └── main.go
│
├── survalentDB/       # Proyecto de C#
│   ├── Program.cs     # Lógica principal y manejo de modos
│   ├── DbTools.cs     # Gestión de la conexión y ejecución de queries
│   ├── JsonStreamer.cs# Serializador de DataTable a JSON
│   └── survalentDB.csproj
│
└── README.md          # Este archivo
```