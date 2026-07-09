# CloseLinkit

## Idea general

En esencia, un acortador de URLs es un servicio que mantiene una relación entre una URL larga y un identificador corto. Cuando alguien visita la URL corta, el servicio consulta esa relación y redirige al navegador a la URL original. 

El flujo típico sería:

```
Usuario
   │
   │ Envía: https://ejemplo.com/productos?id=123&campaña=abc
   ▼
Servidor del acortador
   │
   ├─ Genera un código corto (ej. "A7xP9")
   ├─ Guarda:
   │      A7xP9 → https://ejemplo.com/productos?id=123&campaña=abc
   └─ Devuelve:
          https://miacortador.com/A7xP9
```

Cuando alguien hace clic:

```
https://miacortador.com/A7xP9
             │
             ▼
Servidor del acortador
             │
      Busca "A7xP9" en la base de datos
             │
             ▼
Encuentra la URL original
             │
             ▼
Responde con HTTP 301 o 302
             │
             ▼
El navegador abre:
https://ejemplo.com/productos?id=123&campaña=abc
```

### Componentes principales

- **Base de datos:** almacena la relación `código → URL`.
- **Generador de códigos:** crea identificadores únicos, normalmente usando caracteres alfanuméricos (Base62: `0-9`, `A-Z`, `a-z`).
- **Servidor web:** recibe las peticiones y realiza la redirección HTTP.
- **Panel de estadísticas (opcional):** registra clics, ubicación aproximada, navegador, dispositivo, etc. ([Linkly](https://linklyhq.com/es/blog/url-shortener-system-design))

---

### ¿Cómo se generan los códigos?

Hay varias estrategias:

1. **ID incremental**
    
    ```
    1  → 1
    2  → 2
    3  → 3
    ...
    ```
    
    Luego el número se convierte a Base62:
    
    ```
    125489 → Wf3
    ```
    
2. **Aleatorio**
    
    ```
    Q8xL2
    aP91B
    X7mKd
    ```
    
3. **Hash**
    
    Se calcula un hash de la URL y se utiliza una parte del resultado, verificando que no existan colisiones.
    

La opción más común en servicios grandes es usar un ID único y codificarlo en Base62, porque produce enlaces cortos y evita duplicados con facilidad. ([Wikipedia](https://es.wikipedia.org/wiki/Acortador_de_URL?utm_source=chatgpt.com))

---

### ¿Por qué usar una redirección 301 o 302?

El servidor no envía la página original. En su lugar responde con algo como:

```
HTTP/1.1 301 Moved Permanently
Location: https://ejemplo.com/productos?id=123
```

El navegador recibe esa respuesta y automáticamente solicita la URL indicada.

---

### Funciones adicionales

Muchos acortadores añaden capacidades como:

- estadísticas de clics;
- enlaces personalizados (`miapp.com/oferta`);
- fecha de expiración;
- protección con contraseña;
- códigos QR;
- pruebas A/B o redirecciones según país o dispositivo. ([Shopify](https://www.shopify.com/es/blog/guia-completa-sobre-el-acortador-de-url-y-marketing-en-redes-sociales?utm_source=chatgpt.com))

---

### Escalabilidad

Si el servicio recibe millones de clics diarios, suele incorporar:

- caché (por ejemplo, almacenar las asociaciones más consultadas en memoria);
- bases de datos replicadas;
- balanceadores de carga;
- servidores distribuidos geográficamente para reducir la latencia. ([Linkly](https://linklyhq.com/es/blog/url-shortener-system-design?utm_source=chatgpt.com))

En resumen, un acortador de URLs es conceptualmente sencillo: mantiene un diccionario que asocia un código corto con una URL larga y responde a las solicitudes con una redirección HTTP hacia el destino correspondiente. El reto principal en sistemas de gran escala no es la lógica, sino generar códigos únicos y atender un gran volumen de redirecciones con baja latencia.

---

## v0.1

### Descripción

Permitir que un usuario convierta una URL larga en una URL corta y que cualquier persona pueda utilizar esa URL para ser redirigida al destino original.

Esto mediante una API Rest en Go, usando la librería estándar y por ahora sin frontend, pero posteriormente se implementará uno con TypeScript y React

---

### Funcionalidades

- Fecha de expiración
- Contador de clics

---

### Requisitos funcionales

#### V1

- Crear un enlace corto
- Redirigir al destino
- Consultar información y estadísticas básicas de un enlace

---

### Dominio (entidades)

- Entidad URL
    - id
    - originalURL
    - shortCode
    - createdAt
    - expiresAt
    - clickCount

---

### Flujo general

#### Creación de shortURL

```
Cliente

↓

POST /shorten

↓

Validar URL

↓

Generar código

↓

Guardar

↓

Responder
```

---

#### Redirección

```
Cliente

↓

GET /abc123

↓

Buscar código

↓

Existe?

↓

Sí

↓

301 o 302

↓

URL original
```

---

## Implementaciones futuras

- Implementación de entidad usuario, al autenticarse, esto permitiría:
    - Eliminar urls
    - Crear enlaces personalizados
    - Modificar cosas como hacia adónde apunta un enlace corto o el propio short_code
        - La consulta podría ser algo así
        
        ```sql
        -- name: updateOriginalURL :one
        UPDATE urls
        SET original_url = $1, updated_at = $2
        WHERE short_code = $3
        RETURNING original_url, short_code, created_at, updated_at, expires_at;
        ```
        
    - También podría aumentarse el tiempo de expiración de los enlaces
    - ¿Enlaces privados? (Protección con contraseña)
    - Códigos QR
- Interfaz gráfica de usuario mediante React con TypeScript
- Redirecciones según país o región
- Caché con Redis
- Balanceador de carga
- Probar con alguna herramienta que someta a estrés el servidor
- Más estadísticas como ubicación aproximada, navegador, dispositivo, etc., sitios más frecuentados (quizá solo el dominio?)

---
