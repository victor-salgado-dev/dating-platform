import styles from '../legal.module.css';

export const metadata = { title: 'Aviso Legal' };

export default function ImpressumPage() {
  return (
    <main className={styles.main}>
      <p className={styles.placeholderNotice}>
        <strong>Plantilla, no es asesoramiento legal.</strong> El contenido exacto exigido
        para un aviso legal / Impressum varía mucho según el país (por ejemplo, es un
        requisito específico y detallado en Alemania). Completa los datos reales de tu
        empresa y revísalo con un abogado antes de publicarlo.
      </p>

      <h1>Aviso Legal</h1>

      <h2>Titular del sitio</h2>
      <p>
        [NOMBRE DE LA EMPRESA / PERSONA RESPONSABLE]
        <br />
        [FORMA JURÍDICA, p. ej. S.L. / GmbH]
        <br />
        [DIRECCIÓN COMPLETA]
        <br />
        [PAÍS]
      </p>

      <h2>Contacto</h2>
      <p>
        Email: [EMAIL DE CONTACTO]
        <br />
        Teléfono: [TELÉFONO, si aplica]
      </p>

      <h2>Registro / identificación fiscal</h2>
      <p>
        [PLACEHOLDER: número de registro mercantil, NIF/CIF/VAT ID u otro identificador
        exigido en tu jurisdicción].
      </p>

      <h2>Responsable editorial</h2>
      <p>[PLACEHOLDER: persona responsable del contenido, si la normativa local lo exige].</p>

      <h2>Resolución de litigios</h2>
      <p>
        [PLACEHOLDER: enlace a la plataforma de resolución de litigios en línea de la UE
        u otro mecanismo aplicable, si corresponde].
      </p>
    </main>
  );
}
