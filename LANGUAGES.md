# Idiomas (i18n)

Este proyecto usa una estrategia de internacionalización **propia, sin dependencias externas**, basada en una cookie (`NEXT_LOCALE`) y en diccionarios TypeScript tipados. No hay `middleware.ts`, no hay rutas `app/[locale]/...` y no hay librerías como `next-intl`.

- Idioma por defecto: **Español (`es`)**.
- Idioma secundario actual: **Inglés (`en`)**.
- Persistencia: cookie `NEXT_LOCALE`, `SameSite=Lax`, `Path=/`, 1 año de duración.
- Todo el código de infraestructura vive en `frontend/lib/i18n/`.

---

## 1. Estructura de archivos

