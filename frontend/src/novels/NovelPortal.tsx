import type { Dashboard } from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import type { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import CatalogImage from "../catalog/CatalogImage";
import Icon from "../components/Icon";
import { novelColors } from "./presentation";

export function NovelCover({
  url,
  label,
  lazy = false,
}: {
  url: string;
  label: string;
  lazy?: boolean;
}) {
  return (
    <CatalogImage
      url={url}
      lazy={lazy}
      fallback={
        <div className="cover-placeholder" aria-hidden="true">
          <Icon name="books" size={24} />
          <span className="cover-title">{label}</span>
        </div>
      }
    />
  );
}
export default function NovelPortal({
  dashboard: d,
  info,
  busy,
  onList,
  onNovel,
}: {
  dashboard: Dashboard;
  info: Info;
  busy: boolean;
  onList: (status: string) => void;
  onNovel: (id: string, number?: number) => void;
}) {
  return (
    <div className="novel-portal" role="region" aria-label="Portal de novelas">
      <section className="portal-card novel-library-summary">
        <div className="card-heading">
          <h2>Tu biblioteca</h2>
          <strong className="novel-library-total">
            {d.total}
            <small className="sr-only"> novelas</small>
          </strong>
        </div>
        <div className="novel-state-counts">
          {d.counts.map((count) => (
            <button
              key={count.status}
              disabled={busy}
              onClick={() => onList(count.status)}
            >
              <span
                className="volume-state-dot"
                style={{ background: novelColors[count.status] }}
              />
              <span>
                {
                  info.statuses.find((status) => status.value === count.status)
                    ?.label
                }
              </span>
              <strong>{count.count}</strong>
            </button>
          ))}
        </div>
        <div
          className="novel-distribution"
          aria-label="Distribución de estados"
        >
          {d.counts
            .filter((count) => count.count > 0)
            .map((count) => {
              const label =
                info.statuses.find((status) => status.value === count.status)
                  ?.label || count.status;
              const percent = d.total ? (count.count * 100) / d.total : 0;
              return (
                <button
                  key={count.status}
                  disabled={busy}
                  style={{
                    background: novelColors[count.status],
                    flex: count.count,
                  }}
                  title={`${label}: ${count.count} novelas (${percent.toFixed(1)}%)`}
                  aria-label={`${label}: ${count.count} novelas, ${percent.toFixed(1)}%`}
                  onClick={() => onList(count.status)}
                />
              );
            })}
        </div>
        <button
          className="novel-library-all"
          disabled={busy}
          onClick={() => onList("")}
        >
          <Icon name="books" size={17} />
          <span>Ver todas las novelas</span>
          <Icon name="chevron" size={15} />
        </button>
      </section>
      <section className="portal-card novel-reading-card">
        <div className="card-heading">
          <h2>Leyendo ahora</h2>
          <button
            className="text-action"
            disabled={busy}
            onClick={() => onList("reading")}
          >
            Ver leyendo <Icon name="chevron" size={13} />
          </button>
        </div>
        {!d.reading.length && (
          <p className="management-hint">No hay novelas en lectura.</p>
        )}
        <div className="novel-reading-list">
          {d.reading.map((n) => (
            <button
              key={n.id}
              className="novel-reading-entry"
              disabled={busy}
              aria-label={`Abrir novela ${n.title}`}
              onClick={() => onNovel(n.id)}
            >
              <span className="novel-reading-cover">
                <NovelCover url={n.coverURL} label={n.title} />
              </span>
              <span>
                <strong>{n.title}</strong>
                <small>{n.authors.join(" · ")}</small>
                <small>
                  {n.completed ?? "?"}/{n.total ?? "?"} completados · Volumen{" "}
                  {n.label}
                </small>
              </span>
            </button>
          ))}
        </div>
      </section>
      <section className="portal-card novel-up-next-card">
        <div className="card-heading">
          <h2>A continuación</h2>
          <button
            className="text-action"
            disabled={busy}
            onClick={() => onList("up-next")}
          >
            Ver A continuación <Icon name="chevron" size={13} />
          </button>
        </div>
        <div className="novel-cover-row">
          {d.upNext.map((n) => (
            <button
              key={n.id}
              disabled={busy}
              title={`${n.title} · ${n.number ? `Volumen ${n.label}` : "Progreso desconocido"}`}
              aria-label={`Abrir siguiente ${n.title}`}
              onClick={() => onNovel(n.id)}
            >
              <NovelCover url={n.coverURL} label={n.title} />
            </button>
          ))}
        </div>
        {!d.upNext.length && (
          <p className="management-hint">Tu próxima lectura aparecerá acá.</p>
        )}
      </section>
      <section className="portal-card novel-recent-card">
        <div className="card-heading">
          <h2>Últimos volúmenes leídos</h2>
          <button
            className="text-action"
            disabled={busy}
            onClick={() => onList("read")}
          >
            Ver leídas <Icon name="chevron" size={13} />
          </button>
        </div>
        <div className="novel-cover-row">
          {d.recent.map((v) => (
            <button
              key={`${v.novelID}:${v.number}`}
              disabled={busy}
              title={`${v.novelTitle} · Volumen ${v.label} · ${v.finishedOn}`}
              aria-label={`Abrir ${v.novelTitle}, volumen ${v.label}`}
              onClick={() => onNovel(v.novelID, v.number)}
            >
              <NovelCover url={v.coverURL} label={v.label} />
            </button>
          ))}
        </div>
        {!d.recent.length && (
          <p className="management-hint">
            Las finalizaciones con fecha conocida aparecerán acá.
          </p>
        )}
      </section>
    </div>
  );
}
