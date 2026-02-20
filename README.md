# goScadaSur 🌿

**goScadaSur** es una herramienta avanzada de gestión de base de datos y generación de archivos de configuración para sistemas **SURVALENT SCADA**. Diseñada con una arquitectura moderna que combina la potencia de **Go** en el backend y la interactividad de **React** en el frontend, todo empaquetado siguiendo los lineamientos de la identidad visual de **EPM**.

---

## 🚀 Características Principales

### 🖥️ Interfaz de Usuario (Web UI)
- **Dashboard en Tiempo Real:** Visualización instantánea de estadísticas de plantillas cargadas y estado de conexión con el motor de base de datos.
- **Búsqueda de Estaciones:** Formulario intuitivo para consultar señales directamente en Survalent y visualizar resultados tabulares.
- **Editor de Plantillas Pro:**
  - **Modo Formulario:** Gestión visual de elementos *Analog*, *Discrete* y *Breaker* con búsqueda y clonación rápida.
  - **Modo JSON:** Edición cruda para usuarios avanzados con previsualización en tiempo real.
- **Mapa de Red (DASIP):** Gestión dinámica de mapeos de red para rutas IFS dinámicas.
- **Identidad EPM:** Interfaz diseñada bajo el manual de marca de EPM, utilizando tipografía VAG Rounded y paleta de colores corporativa.

### 🛠️ Capacidades del Motor (Backend)
- **Generación IFS e IMM:** Procesamiento automático de archivos CSV y Excel para generar configuraciones SCADA válidas.
- **Procesamiento Paralelo:** Motor multi-hilos configurable para procesar grandes volúmenes de datos con máxima eficiencia.
- **Interoperabilidad C#:** Comunicación segura vía JSON con herramientas externas para acceso directo a base de datos.
- **Binario Único:** El frontend React se compila y embebe directamente en el ejecutable de Go para una distribución minimalista.

---

## 🛠️ Instalación y Construcción

### Requisitos
- **Go** 1.24+
- **Node.js** & **npm** (solo para desarrollo/compilación)
- **Make**

### Pasos de Compilación
Para generar la distribución completa en la carpeta `dist/`:

```bash
make build
```

Este comando descargará dependencias, compilará el frontend, lo embeberá en el binario de Go y preparará los archivos de configuración necesarios.

---

## 📋 Uso

### Modo Servidor (Interfaz Web)
Para iniciar la aplicación con la interfaz moderna y abrir automáticamente el navegador:

```bash
./dist/goScadaSur serve --port 8080
```

### Modo CLI (Línea de Comandos)
También puedes usar las funciones clásicas directamente desde la terminal:

- **Búsqueda de estación:**
  ```bash
  ./dist/goScadaSur station-search --path EMPRESA/REGION/B1/B2/B3 --aor 107
  ```
- **Conversión CSV a XML:**
  ```bash
  ./dist/goScadaSur csv-xml --path data.csv --aor 107
  ```
- **Query SQL Directa:**
  ```bash
  ./dist/goScadaSur direct-query "SELECT * FROM Table"
  ```

---

## 📂 Estructura del Proyecto

- `cmd/`: Punto de entrada de la aplicación.
- `pkg/api/`: Servidor HTTP REST y manejadores de la UI.
- `pkg/config/`: Gestión de configuración YAML con persistencia.
- `pkg/database/`: Lógica de interoperabilidad con C# y procesamiento de datos.
- `pkg/xmlcreator/`: Motor de transformación de datos a formatos IFS/IMM.
- `web/`: Aplicación frontend en React + TypeScript + Vite.
- `configs/`: Archivos de configuración predeterminados (Templates, DASIP, Config).

---

## 🎨 Identidad Visual
El proyecto implementa los códigos gráficos del **Manual de Marca EPM**:
- **Colores:** Verde Bosque (#0d9648) y Verde Cítrico (#9fcf67).
- **Tipografía:** Thesis Sans y VAG Rounded.
- **Elementos:** Segmentos de circunferencia y formas orgánicas para una experiencia de usuario amigable y profesional.

---
*Desarrollado para la optimización de flujos de trabajo en ingeniería SCADA.*
