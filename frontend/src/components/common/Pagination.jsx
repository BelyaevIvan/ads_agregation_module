export default function Pagination({ total, limit, offset, onChange }) {
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const current = Math.floor(offset / limit) + 1;
  if (totalPages <= 1) return null;

  const go = (page) => {
    const p = Math.min(Math.max(1, page), totalPages);
    onChange((p - 1) * limit);
  };

  const pages = [];
  const add = (n) => pages.push(n);
  const addEllipsis = () => pages.push('…');

  add(1);
  if (current > 3) addEllipsis();
  for (let i = Math.max(2, current - 1); i <= Math.min(totalPages - 1, current + 1); i++) {
    add(i);
  }
  if (current < totalPages - 2) addEllipsis();
  if (totalPages > 1) add(totalPages);

  return (
    <div className="pagination">
      <button className="pg-btn" onClick={() => go(current - 1)} disabled={current === 1}>
        ‹
      </button>
      {pages.map((p, i) =>
        p === '…' ? (
          <span key={`e${i}`} className="pg-btn" style={{ border: 'none' }}>
            …
          </span>
        ) : (
          <button
            key={p}
            className={`pg-btn ${p === current ? 'active' : ''}`}
            onClick={() => go(p)}
          >
            {p}
          </button>
        )
      )}
      <button
        className="pg-btn"
        onClick={() => go(current + 1)}
        disabled={current === totalPages}
      >
        ›
      </button>
    </div>
  );
}
