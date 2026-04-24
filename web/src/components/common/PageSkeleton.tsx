export function PageSkeleton() {
  return (
    <div data-testid="page-skeleton" className="zc" style={{ padding: 24, fontFamily: 'var(--font-sans)' }}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {Array.from({ length: 6 }).map((_, i) => (
          <div
            key={i}
            style={{
              height: 14,
              width: i === 0 ? '40%' : i === 5 ? '70%' : '100%',
              background: 'linear-gradient(90deg, var(--z-100) 0%, var(--z-150) 50%, var(--z-100) 100%)',
              backgroundSize: '200% 100%',
              borderRadius: 4,
              animation: 'zc-skeleton 1.4s ease-in-out infinite',
            }}
          />
        ))}
      </div>
      <style>{`@keyframes zc-skeleton { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }`}</style>
    </div>
  );
}
