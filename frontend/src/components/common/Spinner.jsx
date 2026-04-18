export default function Spinner({ label = 'Загрузка…' }) {
  return (
    <div className="loading-block">
      <div className="spinner" />
      <span>{label}</span>
    </div>
  );
}
