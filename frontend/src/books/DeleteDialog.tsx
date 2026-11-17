import { useEffect, useRef, type KeyboardEvent } from "react";
import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import Icon from "../components/Icon";
import { readingLabels } from "./readingMedia";

export default function DeleteDialog({
  book,
  busy,
  error,
  onCancel,
  onConfirm,
}: {
  book: Book;
  busy: boolean;
  error: string;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  const labels = readingLabels(book.medium);
  const panel = useRef<HTMLDivElement>(null);
  const cancel = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null;
    cancel.current?.focus();
    return () => previous?.focus();
  }, []);
  function handleKeys(event: KeyboardEvent) {
    if (event.key === "Escape" && !busy) {
      event.preventDefault();
      onCancel();
    }
    if (event.key === "Tab") {
      const buttons = panel.current?.querySelectorAll<HTMLButtonElement>(
        "button:not(:disabled)",
      );
      if (!buttons?.length) {
        event.preventDefault();
        return;
      }
      const first = buttons[0],
        last = buttons[buttons.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }
  }
  return (
    <div className="dialog-backdrop">
      <div
        ref={panel}
        className="confirm-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="delete-title"
        aria-describedby="delete-description"
        onKeyDown={handleKeys}
      >
        <span className="delete-symbol">
          <Icon name="trash" size={25} />
        </span>
        <h2 id="delete-title">¿Eliminar este {labels.item}?</h2>
        <p id="delete-description">
          <strong>{book.title}</strong> se quitará de esta biblioteca junto con
          sus notas y registro de lectura. No se eliminarán otros{" "}
          {labels.plural} de su serie. Esta acción no se puede deshacer.
        </p>
        {error && (
          <div className="error-banner" role="alert">
            {error}
          </div>
        )}
        <div className="dialog-actions">
          <button
            className="secondary"
            ref={cancel}
            disabled={busy}
            onClick={onCancel}
          >
            Cancelar
          </button>
          <button className="danger-button" disabled={busy} onClick={onConfirm}>
            {busy ? "Eliminando…" : "Eliminar definitivamente"}
          </button>
        </div>
      </div>
    </div>
  );
}
