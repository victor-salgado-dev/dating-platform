import styles from '../legal.module.css';

export const metadata = { title: 'Política de Privacidad' };

export default function PrivacyPage() {
  return (
    <main className={styles.main}>
      <p className={styles.placeholderNotice}>
        <strong>Plantilla, no es asesoramiento legal.</strong> Este texto describe, a alto
        nivel, qué datos trata realmente esta versión del sistema (para que la
        descripción sea honesta), pero el documento en sí debe ser redactado o revisado
        por un abogado especializado en protección de datos (RGPD u otra normativa
        aplicable) antes de publicarse.
      </p>

      <h1>Política de Privacidad</h1>
      <p className={styles.meta}>
        Versión: v1 · Última actualización: [FECHA] · Responsable: [NOMBRE DE LA EMPRESA], [EMAIL DE CONTACTO]
      </p>

      <h2>1. Qué datos tratamos</h2>
      <ul>
        <li><strong>Cuenta:</strong> email, contraseña (almacenada como hash, nunca en claro), estado de la cuenta.</li>
        <li><strong>Perfil:</strong> nombre visible, fecha de nacimiento, género, país/región, idiomas, objetivo de relación, información familiar, biografía, intereses, fotos.</li>
        <li><strong>Actividad:</strong> favoritos, mensajes enviados/recibidos, bloqueos, reportes.</li>
        <li><strong>Técnicos:</strong> dirección IP y registros de acceso, con fines de seguridad (rate limiting, detección de abuso).</li>
        <li><strong>Consentimientos:</strong> qué versión de estos documentos aceptaste y cuándo (puedes consultarlo en tu <a href="/settings">página de ajustes</a>).</li>
      </ul>
      <p>
        Ningún dato de perfil se asume ni se inventa: si no lo indicas, se guarda como
        &quot;no indicado&quot; y no se trata como si fuera una respuesta negativa.
      </p>

      <h2>2. Datos especialmente sensibles</h2>
      <p>
        Algunos campos de perfil (p. ej. religión, orientación, salud, si se llegaran a
        habilitar en el futuro) pueden constituir categorías especiales de datos según
        la normativa aplicable. [PLACEHOLDER: describir la base legal específica y el
        mecanismo de consentimiento granular para estos campos, si se activan].
      </p>

      <h2>3. Para qué usamos tus datos</h2>
      <ul>
        <li>Prestar el servicio: crear tu perfil, mostrarlo a otros usuarios, permitir la búsqueda y la mensajería.</li>
        <li>Seguridad: prevenir abuso, spam y accesos no autorizados.</li>
        <li>Moderación: revisar reportes y aplicar nuestras normas de conducta.</li>
        <li>Comunicación operativa: verificación de email, recuperación de contraseña.</li>
      </ul>
      <p>No usamos tus datos para publicidad ni los vendemos a terceros.</p>

      <h2>4. Con quién compartimos datos</h2>
      <p>
        Con otros usuarios, en la medida en que tu perfil es visible según las reglas de
        la Plataforma (perfiles bloqueados o de cuentas suspendidas nunca son visibles).
        Con proveedores técnicos estrictamente necesarios para operar el servicio
        (hosting, almacenamiento, envío de email transaccional) — [PLACEHOLDER: listar
        proveedores reales cuando se contraten en producción].
      </p>

      <h2>5. Cuánto tiempo conservamos tus datos</h2>
      <p>
        Mientras tu cuenta esté activa. Al eliminar tu cuenta, se marca como eliminada y
        deja de ser visible o utilizable de inmediato; [PLACEHOLDER: definir el plazo de
        conservación del histórico mínimo antes del borrado físico definitivo, y la base
        legal para conservarlo ese tiempo].
      </p>

      <h2>6. Tus derechos</h2>
      <p>
        Puedes ejercer tus derechos de acceso, rectificación, supresión, portabilidad y
        oposición escribiendo a [EMAIL DE CONTACTO] o desde la{' '}
        <a href="/legal/contact">página de contacto</a>. La eliminación de cuenta está
        disponible en autoservicio desde <a href="/settings">tu página de ajustes</a>.
      </p>

      <h2>7. Seguridad</h2>
      <p>
        Las contraseñas se almacenan con hashing seguro (bcrypt), las sesiones usan
        cookies httpOnly, y aplicamos límites de tasa y validaciones para reducir el
        riesgo de abuso. Ningún sistema es 100% seguro; si detectas una vulnerabilidad,
        repórtala a [EMAIL DE CONTACTO].
      </p>

      <h2>8. Cambios en esta política</h2>
      <p>
        Los cambios sustanciales se comunicarán y, si afectan al tratamiento de tus
        datos, se te pedirá un nuevo consentimiento explícito.
      </p>
    </main>
  );
}
