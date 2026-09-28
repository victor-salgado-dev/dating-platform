import { redirect } from 'next/navigation';

// Ruta antigua: ahora esta lista vive como pestaña dentro de Home.
// Se mantiene solo para no romper enlaces guardados o externos.
export default function LegacyRedirect() {
  redirect('/?tab=popular');
}
