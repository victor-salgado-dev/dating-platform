
@@ -1,4 +1,5 @@
 # Idiomas (i18n)

-Este proyecto usa una estrategia de internacionalización **propia, sin dependencias externas**, basada en una cookie
(`NEXT_LOCALE`) y en diccionarios TypeScript tipados. No hay `middleware.ts`, no hay rutas `app/[locale]/...` y no hay
librerías como `next-intl`.
+Este documento explica cómo funciona el sistema de idiomas del frontend, cómo está implementado, cómo se usa desde el
código y cómo añadir nuevos idiomas además de los que ya existen (**español** e **inglés**).

-- Idioma por defecto: **Español (`es`)**.
-- Idioma secundario actual: **Inglés (`en`)**.
-- Persistencia: cookie `NEXT_LOCALE`, `SameSite=Lax`, `Path=/`, 1 año de duración.
-- Todo el código de infraestructura vive en `frontend/lib/i18n/`.
+El sistema es **propio, sin dependencias externas**. No se usa `next-intl`, `react-i18next`, `i18next` ni ninguna
librería similar. Tampoco hay `middleware.ts`, ni rutas del tipo `app/[locale]/...`, ni reescrituras de URL. Todo se
resuelve con una cookie, un diccionario TypeScript y un Context de React.
+
+- Idioma por defecto: **español (`es`)**.
+- Idiomas disponibles hoy: **`es`** y **`en`**.
+- Persistencia del idioma elegido: cookie `NEXT_LOCALE`.
+- Carpeta donde vive toda la infraestructura: `frontend/lib/i18n/`.

 ---

-## 1. Estructura de archivos
+## 1. Resumen de un vistazo

+- El idioma se guarda en la cookie `NEXT_LOCALE`. Nunca aparece en la URL.
+- Si no hay cookie o su valor no es válido, se cae al idioma por defecto (`es`).
+- Al cambiar de idioma, el `LanguageSwitcher` escribe la cookie y hace `window.location.reload()`. La página se recarga
y el servidor re-renderiza todo el árbol con el diccionario nuevo.
+- Las páginas y componentes que muestran texto leen las cadenas de un objeto `dictionary` tipado. Si una clave falta,
`tsc` lo detecta antes de que llegue a producción.
+- Los diccionarios son `es.ts` y `en.ts`, tipados ambos contra `Dictionary = typeof es`.
+
+---
+
+## 2. Estructura de archivos
+


frontend/lib/i18n/

├── config.ts                      # Locales permitidos, tipo Locale, nombre de la cookie

├── context.tsx                    # I18nProvider + hook useI18n() para Client Components

├── get-dictionary.ts              # getDictionary(locale) para Server Components

├── options.ts                     # Helpers tOption / tOptionList (etiquetas de profileOptions)

├── LanguageSwitcher.tsx           # Botones ES/EN en el footer

└── dictionaries/


├── es.ts                      # Diccionario base — define el tipo Dictionary

└── en.ts                      # Diccionario inglés tipado contra Dictionary



@@ -0,0 +1,51 @@
+
+Ficheros relacionados (fuera de `lib/i18n/`):
+
+- `frontend/app/layout.tsx` — lee la cookie, resuelve el locale, monta `<I18nProvider>` y renderiza el
`<LanguageSwitcher>` en el footer.
+- `frontend/app/legal/*/page.tsx` — Server Components que leen la cookie y usan `generateMetadata()` para que el
`<title>` de la pestaña también cambie con el idioma.
+- `frontend/lib/profileOptions.ts` — contiene solo los valores técnicos (`'female'`, `'long_term'`,
`'importance_intimacy'`, …). Las etiquetas visibles están en el namespace `options` de los diccionarios.
+
+---
+
+## 3. Cómo funciona el cambio de idioma
+
+### 3.1 Piezas que intervienen
+
+1. **Cookie `NEXT_LOCALE`** — almacena la preferencia del usuario. Definida como `LOCALE_COOKIE_NAME` en `config.ts`.
+2. **`config.ts`** — declara qué locales existen (`LOCALES`), cuál es el idioma por defecto (`DEFAULT_LOCALE`) y ofrece
el guard `isValidLocale()`.
+3. **`get-dictionary.ts`** — dada una `Locale` válida, devuelve el objeto `Dictionary` correspondiente.
+4. **`context.tsx`** — un Context de React que expone `locale` y `dictionary` a todo el árbol mediante el hook
`useI18n()`.
+5. **`layout.tsx` (raíz)** — el punto donde todo se conecta: lee la cookie, resuelve el locale, obtiene el diccionario,
asigna `<html lang>` y envuelve el árbol con `<I18nProvider>`.
+6. **`LanguageSwitcher.tsx`** — dos botones (ES / EN) que escriben la cookie y recargan la página.
+
+### 3.2 Flujo completo, paso a paso
+
+**Pintado inicial (lo que ocurre en el servidor):**
+
+1. El usuario entra a cualquier URL (por ejemplo `http://localhost/discover`).
+2. `app/layout.tsx` (Server Component) ejecuta `cookies()` de `next/headers` y lee `NEXT_LOCALE`.
+3. Valida el valor con `isValidLocale()`. Si no es válido o la cookie no existe, usa `DEFAULT_LOCALE` (`'es'`).
+4. Llama a `getDictionary(locale)` y obtiene el diccionario completo del idioma resuelto.
+5. Renderiza `<html lang={locale}>` con el atributo de idioma correcto (útil para lectores de pantalla y para el
traductor del navegador).
+6. Envuelve todo el árbol —`<HeaderChrome>`, la página (`{children}`) y el `<footer>` con el `<LanguageSwitcher>`—
dentro de `<I18nProvider locale={locale} dictionary={dictionary}>`.
+7. El HTML sale del servidor con todos los textos ya en el idioma correcto en el primer pintado (no hay parpadeo de
español a inglés).
+
+**Cambio de idioma (lo que hace el usuario):**
+
+1. El usuario pulsa el botón `EN` (o `ES`) que está en el footer.
+2. `LanguageSwitcher` escribe la cookie: `NEXT_LOCALE=en; Path=/; Max-Age=31536000; SameSite=Lax`.
+3. Inmediatamente después hace `window.location.reload()`.
+4. El navegador pide de nuevo la URL al servidor, esta vez ya con la cookie actualizada.
+5. Vuelve al paso 2 de la sección anterior, ahora con la cookie nueva. Todo el árbol se renderiza con el diccionario
inglés.
+6. Las páginas que tienen `generateMetadata()` (los legales) también actualizan el `<title>` de la pestaña.
+
+### 3.3 Por qué `window.location.reload()` y no `router.refresh()`
+
+En este proyecto se usó inicialmente `router.refresh()`, que es la opción "elegante" de Next. El problema es que
`router.refresh()` **solo invalida el RSC de la ruta actual**: las demás rutas ya visitadas se quedan en el *Client
Router Cache* con el diccionario viejo. Efecto visible: pulsabas `EN` estando en `/settings` y `settings` se traducía,
pero al ir a `/profile` seguía en español hasta recargar a mano.
+
+`window.location.reload()` fuerza un rerender completo desde el servidor. Todas las rutas vuelven a pedir su HTML al
backend y reciben el diccionario correcto. El precio a pagar es que se pierde el estado en memoria (borradores de
formularios, navegación SPA), trade-off aceptable para un selector de idioma que se pulsa muy de vez en cuando.
+
+### 3.4 Sobre `<html lang>`
+
+El atributo `lang` se asigna dinámicamente con el locale resuelto:
+






Impacto:



- Los lectores de pantalla pronunciarán los textos con la fonética correcta.

- Chrome/Safari/Edge ofrecen la traducción automática solo cuando el `lang` no coincide con el idioma del sistema: si un
usuario tiene el navegador en francés y entra con `lang="en"`, aparecerá la barra de "¿Traducir esta página?".



### 3.5 Coste arquitectónico



Leer `cookies()` en el `layout.tsx` raíz hace que **toda la app pase a ser dinámica**. Cada ruta aparece como `ƒ
(Dynamic) server-rendered on demand` en el output de `next build`. Es el precio inevitable de resolver i18n por cookie
sin usar `middleware.ts` ni reescribir las URLs. Como en este proyecto todas las páginas son `'use client'` y piden
datos al backend desde el navegador, no se pierde ninguna ventaja real.



---



## 4. Cómo consumir traducciones desde el código



### 4.1 Client Components (lo más habitual)



Cualquier fichero con `'use client'` en la primera línea puede usar el hook:


'use client';

import { useI18n } from '@/lib/i18n/context';

export default function MyComponent() {

const { locale, dictionary } = useI18n();

// locale: 'es' | 'en'

// dictionary: objeto tipado con todas las claves disponibles

return (


<div>

  <p>{dictionary.common.loading}</p>

  <p>{dictionary.profilePublic.fieldGender}</p>

</div>


);

}


@@ -0,0 +1,7 @@
+
+`useI18n()` lanza un error si se llama fuera del `<I18nProvider>`. Como el provider envuelve todo el árbol desde el
layout raíz, cualquier componente que cuelgue de él lo tiene disponible sin configuración adicional.
+
+### 4.2 Server Components
+
+Un Server Component (sin `'use client'`) no puede usar hooks. Debe leer la cookie él mismo y llamar a
`getDictionary()`:
+


import { cookies } from 'next/headers';

import { DEFAULT_LOCALE, LOCALE_COOKIE_NAME, isValidLocale, type Locale } from '@/lib/i18n/config';

import { getDictionary } from '@/lib/i18n/get-dictionary';

function resolveLocale(): Locale {

const value = cookies().get(LOCALE_COOKIE_NAME)?.value ?? '';

return isValidLocale(value) ? value : DEFAULT_LOCALE;

}

export default function MyServerPage() {

const dictionary = getDictionary(resolveLocale());



return {dictionary.somewhere.title};

}


@@ -0,0 +1,7 @@
+
+Hoy solo los tres documentos legales (`/legal/privacy`, `/legal/terms`, `/legal/impressum`) usan este patrón.
+
+### 4.3 Metadata (`<title>` de la pestaña)
+
+Si el título depende del idioma, se usa `generateMetadata()` async en lugar de `export const metadata`:
+



export async function generateMetadata(): Promise {

const dictionary = getDictionary(resolveLocale());

return { title: dictionary.legal.privacy.title };

}


@@ -0,0 +1,7 @@
+
+El `layout.tsx` raíz todavía tiene metadata estático (`title: 'Dating Platform'`), así que el `<title>` de la pestaña
solo cambia con el idioma en las páginas que implementan `generateMetadata`, hoy únicamente los tres legales.
+
+### 4.4 Interpolación de variables
+
+No hay motor de plantillas. La interpolación se hace con `.replace()` inline en el punto de uso:
+


dictionary.profilePublic.messagePlaceholder.replace('{name}', profile.display_name);

dictionary.common.paginationPage

.replace('{page}', String(page))

.replace('{totalPages}', String(totalPages));


@@ -0,0 +1,3 @@
+
+En el diccionario, el marcador va siempre con la forma `{nombre}`:
+


paginationPage: 'Página {page} de {totalPages}',


@@ -0,0 +1,9 @@
+
+Placeholders que se usan hoy: `{name}`, `{n}`, `{page}`, `{totalPages}`, `{total}`, `{current}`, `{count}`, `{avg}`,
`{version}`, `{region}`, `{country}`, `{label}`.
+
+### 4.5 Etiquetas de `profileOptions.ts`
+
+Los arrays de `frontend/lib/profileOptions.ts` contienen **solo los valores técnicos** que viajan al backend
(`'female'`, `'long_term'`, `'importance_intimacy'`, …). Las etiquetas visibles viven en el namespace `options` de los
diccionarios, con un bucket por categoría cuyo nombre coincide con el del array en camelCase (`options.gender`,
`options.bodyType`, `options.occupation`, `options.languages`, etc.).
+
+Para resolverlas se usan dos helpers:
+


import { tOption, tOptionList } from '@/lib/i18n/options';

tOption(dictionary, 'gender', 'female');                   // → 'Mujer' / 'Woman'

tOptionList(dictionary, 'sports', ['fitness', 'hiking']);  // → 'Fitness, Senderismo' / 'Fitness, Hiking'




Ambos tienen un fallback: si la clave no existe en el diccionario, devuelven el valor crudo (`'female'`) para no romper
la UI si algún día el backend añade un valor nuevo sin que nadie lo registre.



### 4.6 Conmutador de idioma (`LanguageSwitcher.tsx`)



Se renderiza dentro del `<footer>` del layout raíz. Es un Client Component que:



1. Lee el `locale` y `dictionary` del contexto con `useI18n()`.

2. Itera sobre `LOCALES` (importado de `config.ts`) y pinta un botón por idioma. Como el array está tipado, al añadir un
locale nuevo aparece un botón nuevo automáticamente, sin tocar el componente.

3. Al pulsar un botón:

   - Si es el idioma actual, no hace nada.

   - Si es otro, escribe la cookie `NEXT_LOCALE` con `Path=/`, `Max-Age` de un año y `SameSite=Lax`.

   - Hace `window.location.reload()`.

4. Marca el botón activo con `aria-pressed`.



---



## 5. Cómo añadir un nuevo idioma



Ejemplo: añadir **francés** (`fr`). Los cinco pasos son siempre los mismos y no afectan a ninguna otra parte de la app.



### Paso 1 — Registrar el locale en `config.ts`



En `frontend/lib/i18n/config.ts`, añade el código al array `LOCALES`:


export const LOCALES = ['es', 'en', 'fr'] as const;




El tipo `Locale` se infiere automáticamente del array. Al añadir `'fr'`, pasa a ser `'es' | 'en' | 'fr'`. No hay que
tocar `DEFAULT_LOCALE` ni `LOCALE_COOKIE_NAME`.



Efecto inmediato: `isValidLocale('fr')` ahora devuelve `true`.



### Paso 2 — Crear el diccionario



Copia `frontend/lib/i18n/dictionaries/en.ts` a `frontend/lib/i18n/dictionaries/fr.ts`, cambia el nombre de la constante
(`en` → `fr`) y traduce únicamente los **valores**. Las claves se quedan idénticas.


import type { Dictionary } from './es';

export const fr: Dictionary = {

common: {


appName: 'Cercanía',        // marca, se deja igual

loading: 'Chargement…',

// ... resto de claves


},

nav: {


home: 'Accueil',

discover: 'Découvrir',

// ...


},

// ... resto de namespaces

};




Reglas al traducir:



- **No toques las claves**, solo los strings. La forma del objeto debe ser exactamente la de `es.ts`.

- **Conserva los marcadores literalmente**. Si `es.ts` dice `'Foto de {name}'`, en francés será `'Photo de {name}'`, no
`'Photo de Julie'`. El código hace el reemplazo por `{name}` en tiempo de ejecución.

- **Los valores técnicos no se traducen** (`'female'`, `'long_term'`, `'spam'`, `'suspended'`, códigos ISO, URLs,
rutas).

- **Las marcas sí se dejan igual** cuando el equipo lo decide así (`appName: 'Cercanía'`, `Über mich` como título de
sección).



Como el diccionario está tipado contra `Dictionary`, `tsc` te dirá exactamente qué propiedades faltan o sobran. No hay
forma de comitear un diccionario incompleto sin que el build falle.



### Paso 3 — Registrar el diccionario en `get-dictionary.ts`



En `frontend/lib/i18n/get-dictionary.ts`, importa el nuevo diccionario y añádelo al registro:


import { fr } from './dictionaries/fr';

const dictionaries: Record<Locale, Dictionary> = {

es,

en,

fr,

};




El tipo `Record<Locale, Dictionary>` obliga a que todos los locales declarados en `LOCALES` tengan aquí su diccionario.
Si te falta uno, `tsc` te lo dice.



### Paso 4 — Probar en el navegador



El `LanguageSwitcher` itera sobre `LOCALES`, así que el botón `FR` aparecerá solo en el footer, sin que haya que tocar
el componente. Al pulsarlo:



1. Se escribe `NEXT_LOCALE=fr`.

2. La página recarga.

3. El layout raíz lee la cookie, resuelve `'fr'`, pide el diccionario francés y renderiza todo el árbol en francés.

4. El `<html lang>` pasa a `fr`.



### Paso 5 — Verificar el build


cd frontend

cmd /c .\node_modules.bin\tsc --noEmit

cmd /c npm run build




Esperado: 0 errores y `✓ Compiled successfully`. Si el diccionario nuevo no cubre todas las claves, `tsc` listará
exactamente qué propiedades faltan.



---



## 6. Convenciones de claves



Los diccionarios están organizados por **namespace**, no por tipo de string. Una entrada por página o feature:



- `common.*` — strings que aparecen en más de una página (yes/no, close, send, paginación, campos genéricos).

- `nav.*`, `header.*`, `footer.*`, `filters.*`, `languageSwitcher.*` — piezas del shell compartido (header, footer,
barra de filtros).

- `auth.login.*`, `auth.register.*` — login y registro.

- `settings.*` — ajustes de cuenta.

- `profile.*` — vista de mi perfil (`/profile`).

- `profileEdit.*` — formulario de edición (`/profile/edit`).

- `profilePublic.*` — vista de perfil ajeno (`/profiles/[id]`).

- `discover.*`, `home.*` — portada y descubrir.

- `activity.*`, `likes.*`, `visits.*`, `matches.*`, `favorites.*`, `blocked.*`, `search.*` — páginas por feature.

- `messages.*`, `conversation.*` — listado de mensajes y hilo individual.

- `admin.*`, `contact.*` — admin y contacto.

- `legal.*` — términos, privacidad e impressum.

- `options.*` — etiquetas de los valores de `profileOptions.ts`, agrupadas por categoría.

- `reports.*` — razones del formulario de reporte.

- `errors.*` — errores transversales (`errors.rateLimited`).



### Reglas al añadir claves



- **Una clave por cadena lógica**, no por palabra. Las frases compuestas se mantienen enteras para que el traductor
tenga contexto. Si un aviso lleva un `<Link>` en medio, se parte solo lo imprescindible (por ejemplo
`deleteWarningPrefix` + `<Link>` + `deleteWarningSuffix`).

- **No fusiones claves aunque el texto coincida en español.** `Altura` de perfil propio y `Altura` de preferencias de
pareja son claves distintas porque en otros idiomas pueden escribirse distinto.

- **Respeta la forma de `es.ts`.** Los valores son `string` (no `as const`), así que cada idioma puede tener su propio
texto. La estructura debe ser idéntica.

- **Emojis y flechas dentro del string** (`'★ En favoritos'`, `'← Volver a resultados'`), para no fragmentar más de la
cuenta.

- **Marcas y términos técnicos no se traducen**: `appName`, `Über mich` (título de sección), códigos ISO de idioma,
valores de los `<select>` (`'female'`, `'long_term'`, …), rutas y URLs.



### Marcadores de posición



Los placeholders deben conservarse **literalmente** en todos los idiomas; el código los reemplaza por `.replace()`. Si
un traductor los borra o los modifica, la interpolación falla silenciosamente y el marcador aparece tal cual en
pantalla.



---



## 7. Qué NO se traduce (a día de hoy)



Esta i18n cubre solo texto del **frontend**. Quedan fuera, por diseño o por dependencia externa:



- **Textos del backend.** El catálogo de intereses (`InterestDefinition.label`), los enunciados del test de personalidad
(`PersonalityStatement.label`) y los mensajes de error que devuelve la API se muestran tal cual. Los mensajes de error
del frontend en las páginas están traducidos; lo que viene del servidor, no.

- **Valores técnicos que viajan al backend.** `'female'`, `'long_term'`, `'spam'`, `'suspended'`, etc. Son códigos, no
texto.

- **El `metadata` del layout raíz.** Sigue siendo estático (`title: 'Dating Platform'`). El `<title>` de la pestaña solo
cambia en las páginas que implementan `generateMetadata`, hoy únicamente los tres documentos legales.

- **Los `console.error` internos** de depuración.



Si en algún momento hay que traducir el catálogo de intereses o las afirmaciones de personalidad, hay que mover esa
responsabilidad al backend o montar un catálogo traducible con clave (`key → label`) en el frontend. No se puede
resolver solo con los diccionarios actuales.



---



## 8. Notas de arquitectura



- **Sin fallback en runtime a otros locales.** `getDictionary` garantiza a nivel de tipos que el locale recibido es
válido; el `?? dictionaries[DEFAULT_LOCALE]` es solo una red de seguridad.

- **`<html lang>` dinámico.** Cambia con el locale de la cookie; útil para lectores de pantalla y para la traducción
automática del navegador.

- **`useI18n()` no funciona en Server Components.** Lanza error si se usa fuera del provider. En Server Components hay
que llamar a `getDictionary(resolveLocale())` directamente.

- **El build hace todas las rutas dinámicas** por leer `cookies()` en el layout raíz. Es coherente con el resto del
proyecto (todas las páginas son `'use client'` y piden datos al backend desde el navegador).

- **El idioma nunca viaja en la URL.** Es una decisión deliberada: simplifica el enrutado, evita duplicar rutas y hace
que las URLs sean estables para compartir.



---



## 9. Referencias rápidas



- `LOCALES`, `DEFAULT_LOCALE`, `LOCALE_COOKIE_NAME`, `isValidLocale`, `Locale`: `frontend/lib/i18n/config.ts`

- `I18nProvider`, `useI18n`: `frontend/lib/i18n/context.tsx`

- `getDictionary(locale)`, tipo `Dictionary`: `frontend/lib/i18n/get-dictionary.ts`

- `tOption`, `tOptionList`: `frontend/lib/i18n/options.ts`

- Botones ES/EN: `frontend/lib/i18n/LanguageSwitcher.tsx`

- Diccionarios: `frontend/lib/i18n/dictionaries/{es,en}.ts`

- Lectura de la cookie y `<I18nProvider>`: `frontend/app/layout.tsx`

- Ejemplo de `generateMetadata` con locale: `frontend/app/legal/privacy/page.tsx`

- Valores técnicos + etiquetas en diccionario: `frontend/lib/profileOptions.ts` + `options.*`
