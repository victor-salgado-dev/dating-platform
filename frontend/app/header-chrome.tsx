'use client';

import { useCallback, useLayoutEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useAccountData, logoutAndRedirect } from './account-nav';
import { useI18n } from '@/lib/i18n/context';
import styles from './layout.module.css';

// Geometría del bulto (ver también /mnt/user-data/outputs/dating-redesign-mockup.html,
// que es donde se afinaron estos números a base de prueba visual).
const BAR_HEIGHT = 78; // alto de la barra
const BUMP_RADIUS = 66; // radio exterior del bulto (parte de la barra)
const COLLAR = 14; // grosor del collar de color barra visible alrededor del aro
const RIGHT_MARGIN = 56; // separación del borde derecho al punto más a la derecha del bulto

export interface NavLinkItem {
  href: string;
  label: string;
}

export interface FilterLinkItem {
  href: string;
  label: string;
  active?: boolean;
  online?: boolean;
}

export default function HeaderChrome({
  navLinks,
  filterLinks,
}: {
  navLinks: NavLinkItem[];
  filterLinks: FilterLinkItem[];
}) {
  const pathname = usePathname();
  const router = useRouter();
  const account = useAccountData();
  const { dictionary } = useI18n();

  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const [mobileAccOpen, setMobileAccOpen] = useState(false);

  const wrapRef = useRef<HTMLDivElement>(null);
  const shapeRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLElement>(null);
  const pillRef = useRef<HTMLButtonElement>(null);
  const avatarRef = useRef<HTMLDivElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const highlightPillRef = useRef<HTMLDivElement>(null);
  const highlightCircleRef = useRef<HTMLDivElement>(null);

  const topTrackRef = useRef<HTMLElement>(null);
  const topScrollerRef = useRef<HTMLDivElement>(null);
  const filterTrackRef = useRef<HTMLDivElement>(null);
  const filterScrollerRef = useRef<HTMLDivElement>(null);

  // Recorta la barra + el bulto del avatar como UNA sola forma (rectángulo +
  // semicírculo en un solo trazo), y coloca encima el aro/foto y el texto
  // "Mi cuenta". Se recalcula en cada resize para que el bulto no se deforme.
  const updateShape = useCallback(() => {
    const wrap = wrapRef.current;
    const shape = shapeRef.current;
    const pill = pillRef.current;
    const avatar = avatarRef.current;
    const dropdown = dropdownRef.current;
    const hPill = highlightPillRef.current;
    const hCircle = highlightCircleRef.current;
    const content = contentRef.current;
    const bottom = bottomRef.current;
    if (!wrap || !shape) return;

    const W = wrap.clientWidth;
    const cx = W - RIGHT_MARGIN - BUMP_RADIUS;
    const cy = BAR_HEIGHT;

    const d =
      `M0,0 L${W},0 L${W},${BAR_HEIGHT} L${cx + BUMP_RADIUS},${BAR_HEIGHT} ` +
      `A${BUMP_RADIUS},${BUMP_RADIUS} 0 0 1 ${cx - BUMP_RADIUS},${BAR_HEIGHT} L0,${BAR_HEIGHT} Z`;
    shape.style.clipPath = `path('${d}')`;

    if (!pill || !avatar || !dropdown || !hPill || !hCircle) return; // sin sesión: no hay pieza de cuenta que posicionar

    const innerR = BUMP_RADIUS - COLLAR;
    avatar.style.width = `${innerR * 2}px`;
    avatar.style.height = `${innerR * 2}px`;
    avatar.style.left = `${cx - innerR}px`;
    avatar.style.top = `${cy - innerR}px`;

    pill.style.right = `${W - (cx - BUMP_RADIUS) + 6}px`;

    dropdown.style.top = `${cy + BUMP_RADIUS + 22}px`;
    dropdown.style.right = `${W - (cx + BUMP_RADIUS)}px`;

    // Reserva el hueco del nav según la posición REAL del botón, no un valor fijo
    const pillRect = pill.getBoundingClientRect();
    const wrapRect = wrap.getBoundingClientRect();
    const spaceNeeded = Math.round(wrapRect.right - pillRect.left + 24);
    if (content) content.style.paddingRight = `${spaceNeeded}px`;
    if (bottom) bottom.style.paddingRight = `${spaceNeeded}px`;

    // Resaltado fundido (pastilla + círculo como una sola forma al hover/abrir)
    const pillLeftLocal = pillRect.left - wrapRect.left;
    hPill.style.left = `${Math.round(pillLeftLocal - 10)}px`;
    hPill.style.top = `${Math.round(Math.max(BAR_HEIGHT / 2 - 26, cy - BUMP_RADIUS))}px`;
    hPill.style.height = '52px';
    hPill.style.width = `${Math.max(0, Math.round(cx - (pillLeftLocal - 10)))}px`;

    hCircle.style.left = `${cx - BUMP_RADIUS}px`;
    hCircle.style.top = `${cy - BUMP_RADIUS}px`;
    hCircle.style.width = `${BUMP_RADIUS * 2}px`;
    hCircle.style.height = `${BUMP_RADIUS * 2}px`;
  }, []);

  const refreshArrows = useCallback(() => {
    [
      { track: topTrackRef.current, scroller: topScrollerRef.current },
      { track: filterTrackRef.current, scroller: filterScrollerRef.current },
    ].forEach(({ track, scroller }) => {
      if (!track || !scroller) return;
      if (track.scrollWidth > track.clientWidth + 4) {
        scroller.classList.add(styles.overflowing);
      } else {
        scroller.classList.remove(styles.overflowing);
      }
    });
  }, []);

  useLayoutEffect(() => {
    updateShape();
    refreshArrows();
    const handle = () => {
      updateShape();
      refreshArrows();
    };
    window.addEventListener('resize', handle);
    return () => window.removeEventListener('resize', handle);
    // account.photo / authenticated cambian el ancho real del botón -> recalcular
  }, [updateShape, refreshArrows, account.authenticated, account.photo]);

  function scrollTrackBy(ref: React.RefObject<HTMLElement | null>, amount: number) {
    ref.current?.scrollBy({ left: amount, behavior: 'smooth' });
  }

  async function handleLogout() {
    setDropdownOpen(false);
    setMobileAccOpen(false);
    await logoutAndRedirect(router);
  }

  return (
    <>
      {/* ======================= HEADER MÓVIL ======================= */}
      <div className={styles.mobileHeaderBlock}>
        <div className={styles.mobileHeaderTop}>
          <button
            className={styles.mobileHamburger}
            onClick={() => {
              setMobileAccOpen(false);
              setMobileNavOpen((o) => !o);
            }}
            aria-label={dictionary.header.openMenu}
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="21" y2="12" />
              <line x1="3" y1="18" x2="21" y2="18" />
            </svg>
          </button>

          <Link href="/" className={styles.mobileBrand}>
            <span className={styles.mark} />
            {dictionary.common.appName}
          </Link>

          {account.authenticated && (
            <button
              className={styles.mobileAccountBtn}
              style={{
                background: `conic-gradient(${account.completion.color} ${account.completion.percent}%, rgba(255,255,255,0.4) 0)`,
              }}
              onClick={() => {
                setMobileNavOpen(false);
                setMobileAccOpen((o) => !o);
              }}
              aria-label={dictionary.header.myAccount}
            >
              {account.photo ? (
                <img src={account.photo} alt={dictionary.header.avatarAlt} className={styles.mobileAvatarImg} />
              ) : (
                <span className={styles.inner}>{dictionary.header.youFallback}</span>
              )}
            </button>
          )}
          {account.authenticated === false && (
            <Link href="/login" className={styles.mobileLoginBtn}>
              {dictionary.header.loginShort}
            </Link>
          )}
        </div>

        <div className={styles.mobileFilterBar}>
          {filterLinks.map((f) => (
            <Link
              key={f.href + f.label}
              href={f.href}
              className={`${styles.filterBtn} ${f.active ? styles.active : ''}`}
            >
              {f.online && <span className={styles.fdot} />}
              {f.label}
            </Link>
          ))}
        </div>

        <div className={`${styles.mobileDropdown} ${mobileNavOpen ? styles.open : ''}`} style={{ top: 58, left: 14 }}>
          {navLinks.map((n) => (
            <Link
              key={n.href}
              href={n.href}
              className={pathname === n.href ? styles.active : ''}
              onClick={() => setMobileNavOpen(false)}
            >
              {n.label}
            </Link>
          ))}
        </div>

        {account.authenticated && (
          <div className={`${styles.mobileDropdown} ${mobileAccOpen ? styles.open : ''}`} style={{ top: 58, right: 14 }}>
            <Link href="/profile" onClick={() => setMobileAccOpen(false)}>
              {dictionary.header.myProfile}
            </Link>
            <Link href="/settings" onClick={() => setMobileAccOpen(false)}>
              {dictionary.header.settings}
            </Link>
            <button className={styles.item} style={{ color: '#a33' }} onClick={handleLogout}>
              {dictionary.header.logout}
            </button>
          </div>
        )}
      </div>

      {/* ===================== HEADER DE ESCRITORIO ===================== */}
      <div className={styles.headerBlock}>
        <div className={styles.headerShapeWrap} ref={wrapRef}>
          <div className={styles.headerShape} ref={shapeRef} />

          <div className={styles.headerContent} ref={contentRef}>
            <Link href="/" className={styles.brand}>
              <span className={styles.mark} />
              {dictionary.common.appName}
            </Link>

            <div className={styles.navScroller} ref={topScrollerRef}>
              <button className={styles.scrollArrow} onClick={() => scrollTrackBy(topTrackRef, -160)} aria-label={dictionary.header.prev}>
                ‹
              </button>
              <nav className={`${styles.topNav} ${styles.navTrack}`} ref={topTrackRef}>
                {navLinks.map((n) => (
                  <Link key={n.href} href={n.href} className={pathname === n.href ? styles.active : ''}>
                    {n.label}
                  </Link>
                ))}
              </nav>
              <button
                className={`${styles.scrollArrow} ${styles.right}`}
                onClick={() => scrollTrackBy(topTrackRef, 160)}
                aria-label={dictionary.header.next}
              >
                ›
              </button>
            </div>
          </div>

          {account.authenticated && (
            <>
              <div className={styles.accountHighlightPill} ref={highlightPillRef} />
              <div className={styles.accountHighlightCircle} ref={highlightCircleRef} />

              <button className={styles.accountPill} ref={pillRef} onClick={() => setDropdownOpen((o) => !o)}>
                <span>{dictionary.header.myAccount}</span>
              </button>

              <div className={styles.avatarAssembly} ref={avatarRef} onClick={() => setDropdownOpen((o) => !o)}>
                <span
                  className={styles.avatarWrapper}
                  style={{
                    background: `conic-gradient(${account.completion.color} ${account.completion.percent}%, rgba(255,255,255,0.35) 0)`,
                  }}
                >
                  {account.photo ? (
                    <img src={account.photo} alt={dictionary.header.avatarAlt} className={styles.avatarInnerImg} />
                  ) : (
                    <span className={styles.avatarInner}>{dictionary.header.youFallback}</span>
                  )}
                </span>
                <span className={styles.bumpLabel}>{account.completion.percent}%</span>
              </div>

              <div className={`${styles.accountDropdown} ${dropdownOpen ? styles.open : ''}`} ref={dropdownRef}>
                <Link href="/profile" onClick={() => setDropdownOpen(false)}>
                  {dictionary.header.myProfile}
                </Link>
                <Link href="/settings" onClick={() => setDropdownOpen(false)}>
                  {dictionary.header.settings}
                </Link>
                <button className={styles.item} style={{ color: '#a33' }} onClick={handleLogout}>
                  {dictionary.header.logout}
                </button>
              </div>
            </>
          )}

          {account.authenticated === false && (
            <Link href="/login" className={styles.desktopLoginBtn}>
              {dictionary.header.loginLong}
            </Link>
          )}
        </div>

        <nav className={styles.headerBottom} ref={bottomRef}>
          <div className={styles.navScroller} ref={filterScrollerRef}>
            <button className={styles.scrollArrow} onClick={() => scrollTrackBy(filterTrackRef, -160)} aria-label={dictionary.header.prev}>
              ‹
            </button>
            <div className={`${styles.filterTrack} ${styles.navTrack}`} ref={filterTrackRef}>
              {filterLinks.map((f) => (
                <Link
                  key={f.href + f.label}
                  href={f.href}
                  className={`${styles.filterBtn} ${f.active ? styles.active : ''}`}
                >
                  {f.online && <span className={styles.fdot} />}
                  {f.label}
                </Link>
              ))}
            </div>
            <button
              className={`${styles.scrollArrow} ${styles.right}`}
              onClick={() => scrollTrackBy(filterTrackRef, 160)}
              aria-label={dictionary.header.next}
            >
              ›
            </button>
          </div>
        </nav>
      </div>
    </>
  );
}
