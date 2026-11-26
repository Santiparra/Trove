import { useEffect, useState } from "react";
import type {
  Cycle,
  ReadingInput,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import { dateLabel, errorMessage } from "./presentation";
import "./reading-history.css";

export function hasReading(c: Cycle) {
  return !!(
    c.startedOn ||
    c.finishedOn ||
    c.rating != null ||
    ["reading", "read", "paused", "did-not-finish"].includes(c.status)
  );
}

export default function ReadingHistory({
  current,
  history = [],
  info,
  busy,
  onSave,
  onDirty,
  onReread,
  title = "Historial de lecturas",
  allowAdd = true,
  editableCurrent = true,
  labels = {},
}: {
  current: Cycle;
  history?: Cycle[];
  info: Info;
  busy: boolean;
  onSave: (input: ReadingInput) => Promise<boolean>;
  onDirty: (dirty: boolean) => void;
  onReread?: () => void;
  title?: string;
  allowAdd?: boolean;
  editableCurrent?: boolean;
  labels?: Record<string, string>;
}) {
  const [draft, setDraft] = useState<ReadingInput | null>(null);
  const [baseline, setBaseline] = useState("");
  const [error, setError] = useState("");
  const dirty = !!draft && JSON.stringify(draft) !== baseline;
  useEffect(() => {
    onDirty(dirty);
  }, [dirty, onDirty]);
  useEffect(() => () => onDirty(false), [onDirty]);
  const previous = [...history]
    .reverse()
    .sort((a, b) =>
      (b.finishedOn || b.startedOn).localeCompare(a.finishedOn || a.startedOn),
    );
  const readings = [current, ...previous];
  const count = readings.filter(hasReading).length;
  function edit(c?: Cycle) {
    const value: ReadingInput = {
      id: c?.id || "",
      status: c?.status || "read",
      startedOn: c?.startedOn || "",
      finishedOn: c?.finishedOn || "",
      rating: c?.rating ?? null,
      edition: c?.edition ? { ...c.edition } : null,
    };
    setDraft(value);
    setBaseline(JSON.stringify(value));
    setError("");
  }
  function field<K extends keyof ReadingInput>(key: K, value: ReadingInput[K]) {
    setDraft((old) => old && { ...old, [key]: value });
  }
  return (
    <section className="reading-history portal-card" aria-label={title}>
      <div className="card-heading">
        <h2>{title}</h2>
        {onReread &&
          (current.status === "read" ||
            (current.status === "did-not-finish" &&
              history.some((c) => c.status === "read"))) && (
            <button
              className="secondary"
              disabled={busy || !!draft}
              onClick={onReread}
            >
              Releer
            </button>
          )}
      </div>
      <details open={!!draft || undefined}>
        <summary>
          Historial · {count} {count === 1 ? "lectura" : "lecturas"}
          {count > 1
            ? ` · ${count - 1} ${count === 2 ? "relectura" : "relecturas"}`
            : ""}
        </summary>
        <ol className="reading-history-list">
          {readings.map((c, index) => (
            <li key={c.id}>
              <div>
                <strong>
                  {labels[c.id] ||
                    (index === 0
                      ? "Lectura actual / más reciente"
                      : "Lectura anterior")}
                </strong>
                <span>
                  {info.statuses.find((s) => s.value === c.status)?.label ||
                    c.status}{" "}
                  · {dateLabel(c.startedOn)} → {dateLabel(c.finishedOn)}
                </span>
                {c.rating != null && <span>Puntaje: {c.rating}/10</span>}
                {c.edition && (
                  <small>
                    {[
                      c.edition.language,
                      info.formats.find((f) => f.value === c.edition?.format)
                        ?.label,
                      c.edition.identifier,
                    ]
                      .filter(Boolean)
                      .join(" · ")}
                  </small>
                )}
              </div>
              {(index > 0 || editableCurrent) && (
                <button
                  className="quiet"
                  disabled={busy || !!draft}
                  onClick={() => edit(c)}
                  aria-label={`Corregir lectura ${index === 0 ? "actual" : c.id}`}
                >
                  Corregir
                </button>
              )}
            </li>
          ))}
        </ol>
        {allowAdd && (
          <button
            className="secondary"
            disabled={busy || !!draft}
            onClick={() => edit()}
          >
            Agregar lectura pasada
          </button>
        )}
        {draft && (
          <form
            className="reading-history-form"
            onSubmit={async (event) => {
              event.preventDefault();
              setError("");
              try {
                if (await onSave(draft)) {
                  setDraft(null);
                  onDirty(false);
                }
              } catch (reason) {
                setError(errorMessage(reason));
              }
            }}
          >
            <h3>{draft.id ? "Corregir lectura" : "Agregar lectura pasada"}</h3>
            <p className="management-hint">
              Las fechas pueden quedar vacías. Corregir no inicia una relectura
              ni completa fechas automáticamente.
            </p>
            <div className="reading-history-fields">
              <label>
                Estado de esta lectura
                <select
                  disabled={busy}
                  value={draft.status}
                  onChange={(e) => field("status", e.target.value)}
                >
                  {info.statuses
                    .filter(
                      (s) =>
                        draft.id ||
                        ["read", "did-not-finish"].includes(s.value),
                    )
                    .map((s) => (
                      <option key={s.value} value={s.value}>
                        {s.label}
                      </option>
                    ))}
                </select>
              </label>
              <label>
                Inicio de esta lectura
                <input
                  type="date"
                  disabled={busy}
                  value={draft.startedOn}
                  onChange={(e) => field("startedOn", e.target.value)}
                />
              </label>
              <label>
                Fin de esta lectura
                <input
                  type="date"
                  disabled={busy}
                  value={draft.finishedOn}
                  onChange={(e) => field("finishedOn", e.target.value)}
                />
              </label>
              <label>
                Puntaje de esta lectura
                <input
                  type="number"
                  min="1"
                  max="10"
                  step="0.25"
                  disabled={busy}
                  value={draft.rating ?? ""}
                  placeholder="Sin puntaje"
                  onChange={(e) =>
                    field(
                      "rating",
                      e.target.value === "" ? null : Number(e.target.value),
                    )
                  }
                />
              </label>
              {allowAdd && (
                <>
                  <label>
                    Idioma leído
                    <input
                      disabled={busy}
                      value={draft.edition?.language || ""}
                      onChange={(e) =>
                        field("edition", {
                          language: e.target.value,
                          format: draft.edition?.format || "",
                          identifier: draft.edition?.identifier || "",
                        })
                      }
                    />
                  </label>
                  <label>
                    Formato leído
                    <select
                      disabled={busy}
                      value={draft.edition?.format || ""}
                      onChange={(e) =>
                        field("edition", {
                          language: draft.edition?.language || "",
                          format: e.target.value,
                          identifier: draft.edition?.identifier || "",
                        })
                      }
                    >
                      {info.formats.map((f) => (
                        <option key={f.value} value={f.value}>
                          {f.label}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label>
                    ISBN / identificador leído
                    <input
                      disabled={busy}
                      value={draft.edition?.identifier || ""}
                      onChange={(e) =>
                        field("edition", {
                          language: draft.edition?.language || "",
                          format: draft.edition?.format || "",
                          identifier: e.target.value,
                        })
                      }
                    />
                  </label>
                </>
              )}
            </div>
            {error && <p role="alert">{error}</p>}
            <div className="catalog-actions">
              <button type="submit" className="primary" disabled={busy}>
                Guardar lectura
              </button>
              <button
                type="button"
                className="quiet"
                disabled={busy}
                onClick={() => {
                  setDraft(null);
                  onDirty(false);
                }}
              >
                Cancelar
              </button>
            </div>
          </form>
        )}
      </details>
    </section>
  );
}
