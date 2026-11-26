import { useEffect, useRef, type KeyboardEvent, type ReactNode } from "react";
import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import Icon from "../components/Icon";
import { readingLabels } from "./readingMedia";

export default function DeleteDialog({
  book,
  busy,
  error,
  onCancel,
  onConfirm,
  heading,
  description,
  confirmLabel,
  danger = true,
}: {
  book: Book;
  busy: boolean;
  error: string;
  onCancel: () => void;
  onConfirm: () => void;
  heading?: string;
  description?: ReactNode;
  confirmLabel?: string;
  danger?: boolean;
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
        <span className={danger ? "delete-symbol" : "complete-symbol"}>
          <Icon name={danger ? "trash" : "check"} size={25} />
        </span>
        <h2 id="delete-title">{heading || `¿Eliminar este ${labels.item}?`}</h2>
        <p id="delete-description">
          {description || (
            <>
              <strong>{book.title}</strong> se quitará de esta biblioteca junto
              con sus notas y registro de lectura. No se eliminarán otros{" "}
              {labels.plural} de su serie. Esta acción no se puede deshacer.
            </>
          )}
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
          <button
            className={danger ? "danger-button" : "primary"}
            disabled={busy}
            onClick={onConfirm}
          >
            {busy ? "Procesando…" : confirmLabel || "Eliminar definitivamente"}
          </button>
        </div>
      </div>
    </div>
  );
}
