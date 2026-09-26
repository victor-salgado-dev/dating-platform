'use client';

// Empty state y banner de error compartidos. Antes cada pantalla los
// repetía como JSX suelto con sus propias clases; aquí solo se pasa el
// texto y las clases ya existentes de cada CSS module, así que el estilo
// visual no cambia en ningún sitio, solo se deja de duplicar el markup.

export function EmptyState({
  title,
  subtitle,
  className,
  titleClassName,
  subtitleClassName,
}: {
  title: string;
  subtitle?: string;
  className?: string;
  titleClassName?: string;
  subtitleClassName?: string;
}) {
  return (
    <div className={className}>
      <p className={titleClassName}>{title}</p>
      {subtitle && <p className={subtitleClassName}>{subtitle}</p>}
    </div>
  );
}

export function ErrorBanner({ message, className }: { message: string; className?: string }) {
  return <div className={className}>{message}</div>;
}
