import styles from '../legal.module.css';

export const metadata = { title: 'Términos y Condiciones' };

export default function TermsPage() {
  return (
    <main className={styles.main}>
      <p className={styles.placeholderNotice}>
        <strong>Plantilla, no es asesoramiento legal.</strong> Este texto es un punto de
        partida estructural para que un abogado lo revise y adapte a tu jurisdicción,
        tu empresa y tu operativa real antes de publicarlo. No uses esta página tal
        cual en producción.
      </p>

      <h1>Términos y Condiciones</h1>
      <p className={styles.meta}>
        Versión: v1 · Última actualización: [FECHA] · Aplicable desde el registro en la plataforma.
      </p>

      <h2>1. Quiénes somos</h2>
      <p>
        [NOMBRE DE LA EMPRESA], con domicilio en [DIRECCIÓN], [PAÍS] (en adelante, &quot;la
        Plataforma&quot;), opera este servicio de dating/relaciones. Puedes contactarnos en{' '}
        [EMAIL DE CONTACTO] o a través de la <a href="/legal/contact">página de contacto</a>.
      </p>

      <h2>2. Requisitos de edad</h2>
      <p>
        La Plataforma está reservada a personas mayores de 18 años. Al registrarte,
        declaras y garantizas que tienes al menos 18 años. Nos reservamos el derecho de
        suspender o eliminar cualquier cuenta que incumpla este requisito.
      </p>

      <h2>3. Tu cuenta</h2>
      <p>
        Eres responsable de mantener la confidencialidad de tu contraseña y de toda la
        actividad que ocurra en tu cuenta. Debes proporcionar información veraz al
        registrarte y mantenerla actualizada.
      </p>

      <h2>4. Normas de conducta</h2>
      <p>Al usar la Plataforma, te comprometes a no:</p>
      <ul>
        <li>Suplantar la identidad de otra persona ni crear perfiles falsos.</li>
        <li>Acosar, amenazar o intimidar a otros usuarios.</li>
        <li>Publicar contenido ilegal, difamatorio o sexualmente explícito no consentido.</li>
        <li>Usar la Plataforma con fines comerciales no autorizados (spam, publicidad).</li>
        <li>Intentar acceder a cuentas ajenas o vulnerar la seguridad del servicio.</li>
      </ul>
      <p>
        El incumplimiento puede dar lugar a la suspensión o eliminación de tu cuenta,
        conforme a nuestros mecanismos de moderación (bloqueo, reportes y revisión por
        el equipo de la Plataforma).
      </p>

      <h2>5. Contenido del usuario</h2>
      <p>
        Mantienes la propiedad del contenido que publicas (fotos, biografía, mensajes).
        Nos concedes una licencia limitada para almacenarlo y mostrarlo dentro de la
        Plataforma con el único fin de prestar el servicio.
      </p>

      <h2>6. Eliminación de cuenta</h2>
      <p>
        Puedes eliminar tu cuenta en cualquier momento desde{' '}
        <a href="/account">tu página de cuenta</a>. Consulta la{' '}
        <a href="/legal/privacy">Política de Privacidad</a> para saber qué ocurre con tus
        datos tras la eliminación.
      </p>

      <h2>7. Limitación de responsabilidad</h2>
      <p>
        [PLACEHOLDER: cláusula de limitación de responsabilidad específica de tu
        jurisdicción — revisar con un abogado].
      </p>

      <h2>8. Ley aplicable</h2>
      <p>
        Estos términos se rigen por las leyes de [JURISDICCIÓN]. [PLACEHOLDER: fuero
        competente / arbitraje, según corresponda].
      </p>

      <h2>9. Cambios en estos términos</h2>
      <p>
        Podemos actualizar estos términos. Los cambios sustanciales se comunicarán y
        requerirán una nueva aceptación; el sistema conserva un registro con la versión
        y fecha exactas que cada usuario aceptó.
      </p>
    </main>
  );
}
