# Idiomas (i18n)

Este documento explica cómo funciona el sistema de idiomas del frontend, cómo está implementado, cómo se usa desde el código y cómo añadir nuevos idiomas además de los que ya existen (**español** e **inglés**).

El sistema es **propio, sin dependencias externas**. No se usa `next-intl`, `react-i18next`, `i18next` ni ninguna librería similar. Tampoco hay `middleware.ts`, ni rutas del tipo `app/[locale]/...`, ni reescrituras de URL. Todo se resuelve con una cookie, un diccionario TypeScript y un Context de React.

- Idioma por defecto: **español (`es`)**.
- Idiomas disponibles hoy: **`es`** y **`en`**.
- Persistencia del idioma elegido: cookie `NEXT_LOCALE`.
- Carpeta donde vive toda la infraestructura: `frontend/lib/i18n/`.

---

## 1. Resumen de un vistazo

- El idioma se guarda en la cookie `NEXT_LOCALE`. Nunca aparece en la URL.
- Si no hay cookie o su valor no es válido, se cae al idioma por defecto (`es`).
- Al cambiar de idioma, el `LanguageSwitcher` escribe la cookie y hace `window.location.reload()`. La página se recarga y el servidor re-renderiza todo el árbol con el diccionario nuevo.
- Las páginas y componentes que muestran texto leen las cadenas de un objeto `dictionary` tipado. Si una clave falta, `tsc` lo detecta antes de que llegue a producción.
- Los diccionarios son `es.ts` y `en.ts`, tipados ambos contra `Dictionary = typeof es`.

---

## 2. Estructura de archivos

