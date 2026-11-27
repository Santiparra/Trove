import { useEffect, useState } from "react";
import type {
  Input,
  Set as ClassificationSet,
  Term,
} from "../../bindings/github.com/Santiparra/Trove/internal/classification/models";
import "./classification.css";

export default function ClassificationEditor({
  title,
  manual,
  effective,
  busy = false,
  onSave,
  onDirty,
}: {
  title: string;
  manual?: Input | null;
  effective?: ClassificationSet | null;
  busy?: boolean;
  onSave?: (input: Input) => Promise<boolean>;
  onDirty?: (dirty: boolean) => void;
}) {
  const [editing, setEditing] = useState(false);
  const [kind, setKind] = useState("genre");
  const [draft, setDraft] = useState("");
  useEffect(() => {
    onDirty?.(!!draft.trim());
  }, [draft, onDirty]);
  useEffect(() => () => onDirty?.(false), [onDirty]);
  const input = (): Input => ({
    genres: [...(manual?.genres || [])],
    tags: [...(manual?.tags || [])],
    hidden: [...(manual?.hidden || [])],
  });
  async function remove(term: Term, type: string) {
    const next = input();
    const field = type === "genre" ? "genres" : "tags";
    next[field] = next[field].filter(
      (name) => name.toLowerCase() !== term.name.toLowerCase(),
    );
    for (const origin of term.origins) {
      if (origin.provider !== "manual")
        next.hidden.push({
          kind: type,
          provider: origin.provider,
          id: origin.id,
        });
    }
    await onSave?.(next);
  }
  return (
    <section className="classification-editor" aria-label={title}>
      <h3>{title}</h3>
      {[
        { key: "genre", label: "Géneros", terms: effective?.genres || [] },
        { key: "tag", label: "Etiquetas", terms: effective?.tags || [] },
      ].map((group) => (
        <div className="classification-row" key={group.key}>
          <span className="classification-label">{group.label}</span>
          <div className="classification-terms">
            {group.terms.length ? (
              group.terms.map((term, index) => (
                <span
                  className="classification-chip"
                  key={`${index}-${term.name}`}
                  title={term.origins
                    .map((o) =>
                      o.provider === "manual"
                        ? "Manual"
                        : `${o.provider} · ${o.category} · ${o.id}`,
                    )
                    .join("\n")}
                >
                  {term.name}
                  {onSave && editing && (
                    <button
                      type="button"
                      className="quiet"
                      disabled={busy}
                      aria-label={`Quitar ${group.key === "genre" ? "género" : "etiqueta"} ${term.name} de ${title}`}
                      onClick={() => void remove(term, group.key)}
                    >
                      ×
                    </button>
                  )}
                </span>
              ))
            ) : (
              <span className="classification-empty">Sin especificar</span>
            )}
          </div>
        </div>
      ))}
      {onSave && (
        <details onToggle={(event) => setEditing(event.currentTarget.open)}>
          <summary>Editar clasificación</summary>
          <form
            onSubmit={async (event) => {
              event.preventDefault();
              const next = input();
              next[kind === "genre" ? "genres" : "tags"].push(draft);
              if (await onSave(next)) {
                setDraft("");
                onDirty?.(false);
              }
            }}
          >
            <select
              aria-label={`Categoría para ${title}`}
              disabled={busy}
              value={kind}
              onChange={(e) => setKind(e.target.value)}
            >
              <option value="genre">Género</option>
              <option value="tag">Etiqueta</option>
            </select>
            <input
              aria-label={`Nuevo término para ${title}`}
              disabled={busy}
              maxLength={80}
              required
              value={draft}
              placeholder={
                kind === "genre" ? "Agregar género…" : "Agregar etiqueta…"
              }
              onChange={(e) => setDraft(e.target.value)}
            />
            <button className="secondary" disabled={busy || !draft.trim()}>
              Agregar
            </button>
          </form>
          {!!manual?.hidden.length && (
            <button
              type="button"
              className="quiet"
              disabled={busy}
              onClick={() => void onSave({ ...input(), hidden: [] })}
            >
              Restaurar términos del proveedor ({manual.hidden.length})
            </button>
          )}
          <small>
            Los cambios manuales se conservan al refrescar el catálogo. Quitar
            un término del proveedor lo oculta hasta restaurarlo.
          </small>
        </details>
      )}
    </section>
  );
}
