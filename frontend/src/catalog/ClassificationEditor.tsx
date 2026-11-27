import { useEffect, useRef, useState } from "react";
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
  const dialogRef = useRef<HTMLDialogElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    const dialog = dialogRef.current;
    if (editing && dialog && !dialog.open) {
      if (typeof dialog.showModal === "function") dialog.showModal();
      else dialog.setAttribute("open", "");
      dialog.querySelector("input")?.focus();
    }
  }, [editing]);
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
  function rows(editable: boolean) {
    return [
      { key: "genre", label: "Géneros", terms: effective?.genres || [] },
      { key: "tag", label: "Etiquetas", terms: effective?.tags || [] },
    ].map((group) => (
      <div className={`classification-row is-${group.key}`} key={group.key}>
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
                {editable && (
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
    ));
  }
  function close() {
    if (busy) return;
    const dialog = dialogRef.current;
    if (typeof dialog?.close === "function") dialog.close();
    setDraft("");
    setEditing(false);
    trigger.current?.focus();
  }
  return (
    <section className="classification-editor" aria-label={title}>
      <div className="classification-heading">
        <h3>{title}</h3>
        {onSave && (
          <button
            ref={trigger}
            className="classification-add"
            type="button"
            aria-label={`Editar clasificación: ${title}`}
            title="Agregar o editar géneros y etiquetas"
            disabled={busy}
            onClick={() => setEditing(true)}
          >
            +
          </button>
        )}
      </div>
      <div className="classification-display">{rows(false)}</div>
      {onSave && editing && (
        <dialog
          ref={dialogRef}
          className="classification-dialog"
          aria-label={`Editar clasificación: ${title}`}
          onCancel={(event) => {
            event.preventDefault();
            close();
          }}
        >
          <div className="classification-heading">
            <h3>{title}</h3>
            <button
              type="button"
              className="quiet"
              aria-label="Cerrar clasificación"
              disabled={busy}
              onClick={close}
            >
              ×
            </button>
          </div>
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
          {rows(true)}
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
        </dialog>
      )}
    </section>
  );
}
