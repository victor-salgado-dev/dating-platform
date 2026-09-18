Hola, estoy desarrollando una plataforma internacional de citas construida con Next.js (App Router, TypeScript), Go, PostgreSQL y Docker Compose.

Ya hemos completado, migrado y probado al 100% el backend (Base de Datos + API REST en Go + Buscador). Ahora necesitamos implementar la **PARTE 4: El Frontend (Next.js)** para la edición del perfil (`frontend/app/profile/edit/page.tsx`), asegurando que todos los campos sean fijos (desplegables/checkboxes, sin texto libre salvo bio y nombre) para que el buscador funcione con precisión.

### 1. Estado y Reglas del Proyecto:
- **Regla de datos faltantes:** Lo que el usuario no seleccione debe enviarse como `null` a la API (nunca strings vacíos ni valores inventados).
- **Frontend actual:** Usa `apiFetch` desde `@/lib/api` contra `/api/v1/profiles/me` (métodos `POST` al crear y `PATCH` al actualizar).
- **Diseño del frontend:** Para evitar que `page.tsx` sea inmanejable, queremos separar los catálogos y opciones en un archivo de constantes (`frontend/lib/profile-options.ts`) e importarlos en el formulario.

### 2. Contrato de la API y Campos Exactos en el Backend:
La API espera y devuelve los siguientes campos estructurados en JSON:

#### A. Datos Básicos y Existentes:
- `display_name` (string, obligatorio)
- `birth_date` (string "YYYY-MM-DD", obligatorio solo en POST/creación)
- `gender` (select: 'female', 'male', 'non_binary', 'other')
- `country_code` (select de países con código ISO 2 letras: 'ES', 'MX', 'US', etc.)
- `region` (string o null)
- `languages` (multi-select / array de códigos ISO: 'es', 'en', 'fr', 'de', 'it', 'pt', etc.)
- `relationship_goal` (select: 'casual', 'long_term', 'friendship', 'marriage', 'not_sure')
- `has_children` (select: 'yes', 'no', 'prefer_not_to_say') [CAMBIADO DE BOOLEAN A TEXTO]
- `wants_children` (select: 'yes', 'no', 'not_sure') [CAMBIADO DE BOOLEAN A TEXTO]
- `bio` (textarea libre, máx 1000 caracteres)
- `interests` (array de strings)

#### B. Físico y Apariencia:
- `height` (número entero en cm, ej: 172)
- `weight` (número entero en kg, ej: 65)
- `body_type` (select: 'petite', 'slim', 'athletic', 'average', 'few_extra_pounds', 'full_figured', 'large_and_lovely')
- `ethnicity` (select: 'arab', 'asian', 'black', 'caucasian', 'hispanic', 'indian', 'mixed', 'pacific_islander', 'other')
- `appearance_rating` (select: 'below_average', 'average', 'attractive', 'very_attractive')
- `hair_color` (select: 'bald', 'black', 'blonde', 'brown', 'grey', 'light_brown', 'red', 'changes_frequently', 'other')
- `eye_color` (select: 'black', 'blue', 'brown', 'green', 'grey', 'hazel', 'other')
- `body_art` (multi-select / array: 'branding', 'earrings', 'piercing', 'tattoo', 'other', 'none')

#### C. Estilo de Vida y Familia:
- `smoking_habit` (select: 'yes', 'no', 'occasionally')
- `drinking_habit` (select: 'yes', 'no', 'occasionally')
- `relocation_willingness` (multi-select / array: 'within_country', 'another_country', 'not_willing', 'not_sure')
- `marital_status` (select: 'single', 'separated', 'widowed', 'divorced', 'other')
- `children_count` (número entero o null)
- `youngest_child_age` (número entero o null)
- `oldest_child_age` (número entero o null)
- `occupation` (select con opciones estándar: 'administrative', 'advertising', 'artistic', 'construction', 'domestic_helper', 'education', 'entertainment', 'executive', 'farming', 'finance', 'fire_law_enforcement', 'hair_dresser', 'it_communications', 'laborer', 'legal', 'medical', 'military', 'nanny', 'none', 'non_profit', 'political', 'retail', 'retired', 'sales', 'self_employed', 'sports', 'student', 'technical', 'transportation', 'travel', 'unemployed', 'other')
- `employment_status` (select: 'student', 'part_time', 'full_time', 'homemaker', 'retired', 'not_employed', 'other')
- `income_level` (select: 'low', 'medium', 'high', 'very_high', 'prefer_not_to_say')
- `living_situation` (select: 'live_alone', 'live_with_friends', 'live_with_family', 'live_with_kids', 'live_with_spouse', 'other')

#### D. Cultura, Fondo y Valores:
- `nationality` (select código ISO 2 letras, ej: 'ES', 'FR')
- `education_level` (select: 'high_school', 'associates', 'bachelors', 'masters', 'phd', 'other')
- `english_ability` (select: 'none', 'basic', 'intermediate', 'fluent', 'native')
- `religion` (select: 'bahai', 'buddhist', 'catholic', 'christian_other', 'protestant', 'hindu', 'islam', 'jainism', 'jewish', 'parsi', 'shintoism', 'sikhism', 'taoism', 'other', 'none')
- `religious_values` (select: 'not_religious', 'religious', 'very_religious')
- `star_sign` (select: 'aquarius', 'aries', 'cancer', 'capricorn', 'gemini', 'leo', 'libra', 'pisces', 'sagittarius', 'scorpio', 'taurus', 'virgo')

### 3. Objetivo actual:
Quiero actualizar el Frontend en dos pasos:
1. Crear el archivo `frontend/lib/profile-options.ts` con todos los catálogos y etiquetas en español listos para usar en desplegables.
2. Actualizar el tipado en `frontend/lib/api.ts` y la vista `frontend/app/profile/edit/page.tsx` para estructurar el formulario de manera limpia, moderna y por secciones legibles.