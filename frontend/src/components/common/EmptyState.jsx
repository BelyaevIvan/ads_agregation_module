export default function EmptyState({ icon = '📭', title, subtitle }) {
  return (
    <div className="empty-state">
      <div className="empty-state-icon">{icon}</div>
      {title && <div className="empty-state-title">{title}</div>}
      {subtitle && <div>{subtitle}</div>}
    </div>
  );
}
