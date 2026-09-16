'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';

import { apiFetch, PublicProfile, ProfilePhoto } from '@/lib/api';
import styles from './layout.module.css';

export default function AccountNav() {
  const pathname = usePathname();
  const router = useRouter();
  
  const [authenticated, setAuthenticated] = useState<boolean | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);
  
  // Datos para el avatar
  const [myPhoto, setMyPhoto] = useState<string | null>(null);
  const [completion, setCompletion] = useState({ percent: 0, color: '#e0e0e0' });

  useEffect(() => {
    let active = true;

    setAuthenticated(null);
    apiFetch('/auth/me')
      .then(() => {
        if (!active) return;
        setAuthenticated(true);
        
        // Si estamos logueados, traemos el perfil y la foto
        Promise.all([
          apiFetch<PublicProfile>('/profiles/me').catch(() => null),
          apiFetch<ProfilePhoto[]>('/profiles/me/photos').catch(() => [])
        ]).then(([profile, photos]) => {
          if (!active) return;

          // 1. Establecer Foto
          if (photos && photos.length > 0) {
            setMyPhoto(photos[0].url ?? `/api/v1/profiles/me/photos/${photos[0].id}/file`);
          }

          // 2. Calcular % de completado
          let score = 0;
          let maxScore = 10; // Criterios totales
          
          if (profile) {
            score += 4; // Datos base (nombre, genero, fecha nac, pais) siempre existen si hay perfil
            if (profile.region) score++;
            if (profile.languages && profile.languages.length > 0) score++;
            if (profile.relationship_goal) score++;
            if (profile.has_children !== null) score++;
            if (profile.bio) score++;
            if (profile.interests && profile.interests.length > 0) score++;
          }
          if (photos && photos.length > 0) score += 2; // Extra por foto
          
          maxScore += 2; // total 12
          const percent = Math.round((score / maxScore) * 100);
          
          // Lógica de colores (Rojo -> Amarillo -> Verde)
          let color = '#ef4444'; // Rojo por defecto
          if (percent >= 50 && percent < 80) color = '#eab308'; // Amarillo
          if (percent >= 80) color = '#22c55e'; // Verde

          setCompletion({ percent, color });
        });

      })
      .catch(() => {
        if (active) setAuthenticated(false);
      });

    return () => {
      active = false;
    };
  }, [pathname]);

  useEffect(() => {
    function handleAuthChange() {
      setAuthenticated(false);
    }
    window.addEventListener('auth-change', handleAuthChange);
    return () => window.removeEventListener('auth-change', handleAuthChange);
  }, []);

  async function handleLogout() {
    try {
      await apiFetch<void>('/auth/logout', { method: 'POST' });
    } catch {}
    setAuthenticated(false);
    setMenuOpen(false);
    window.dispatchEvent(new Event('auth-change'));
    router.push('/login');
  }

  if (authenticated === null) return null;

  if (!authenticated) {
    return (
      <Link href="/login" style={{ fontWeight: 'bold', border: '1px solid white', padding: '0.5rem 1rem', borderRadius: '4px' }}>
        Iniciar sesión
      </Link>
    );
  }

  // Estilo dinámico para el anillo de progreso
  const ringStyle = {
    background: `conic-gradient(${completion.color} ${completion.percent}%, #e0e0e0 0)`
  } as React.CSSProperties;

  return (
    <div className={styles.accountMenu}>
      <button
        type="button"
        className={styles.accountButton}
        onClick={() => setMenuOpen((open) => !open)}
        aria-expanded={menuOpen}
      >
        <span style={{ marginRight: '8px' }}>Mi Cuenta ▼</span>
        <div className={styles.avatarWrapper} style={ringStyle}>
          {myPhoto ? (
            <img src={myPhoto} alt="Mi Avatar" className={styles.avatarInner} />
          ) : (
            <div className={styles.avatarInner} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f3f4f6', color: '#9ca3af', fontSize: '10px', textAlign: 'center', lineHeight: '1' }}>
              Sin Foto
            </div>
          )}
        </div>
      </button>

      {menuOpen && (
        <div className={styles.accountDropdown} role="menu">
          <Link href="/profile" role="menuitem" onClick={() => setMenuOpen(false)}>
            👤 Mi perfil
          </Link>
          <Link href="/settings" role="menuitem" onClick={() => setMenuOpen(false)}>
            ⚙️ Ajustes
          </Link>
          <button type="button" role="menuitem" onClick={handleLogout} style={{ color: '#ef4444', borderTop: '1px solid #eee' }}>
            🚪 Cerrar sesión
          </button>
        </div>
      )}
    </div>
  );
}