import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Info } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import Icon from "../components/Icon";
import BookCover from "./BookCover";
import StatusSelect from "./StatusSelect";
import { dateLabel, labelFor, seriesLabel } from "./presentation";

export default function BookDetails({
  book,
  info,
  busy,
  onBack,
  onEdit,
  onDelete,
  onStatus,
}: {
  book: Book;
  info: Info;
  busy: boolean;
  onBack: () => void;
  onEdit: () => void;
  onDelete: (book: Book) => void;
  onStatus: (book: Book, status: string) => void;
}) {
  return (
    <section className="book-page" aria-label="Detalle del libro">
      <button className="back-link" onClick={onBack} disabled={busy}>
        <Icon name="arrow" size={17} /> Volver
      </button>
      <div className="book-hero">
        <div className="hero-cover">
          <BookCover book={book} />
        </div>
        <div className="book-hero-text">
          <p className="eyebrow">EN TU BIBLIOTECA · REGISTRO LOCAL</p>
          <h1>{book.title}</h1>
          <p className="hero-author">
            {book.authors.join(" · ") || "Autoría sin especificar"}
          </p>
          {book.series && <p className="hero-series">{seriesLabel(book)}</p>}
          <div className="book-actions">
            <StatusSelect
              book={book}
              options={info.statuses}
              disabled={busy}
              onChange={onStatus}
            />
            <button className="secondary" onClick={onEdit} disabled={busy}>
              <Icon name="edit" size={16} /> Editar libro
            </button>
            <button
              className="quiet danger-text"
              onClick={() => onDelete(book)}
              disabled={busy}
            >
              <Icon name="trash" size={16} /> Eliminar libro
            </button>
          </div>
          <p className="local-hint">
            Podés cambiar el estado directamente. Las fechas se completan desde
            Editar.
          </p>
        </div>
      </div>
      <div className="book-detail-grid">
        <section className="portal-card">
          <div className="card-heading">
            <h2>Tu lectura</h2>
          </div>
          <dl className="book-facts">
            <dt>Estado</dt>
            <dd>{labelFor(info.statuses, book.cycle.status)}</dd>
            <dt>Añadido</dt>
            <dd>{dateLabel(book.addedOn)}</dd>
            <dt>Inicio</dt>
            <dd>{dateLabel(book.cycle.startedOn)}</dd>
            <dt>Finalización</dt>
            <dd>{dateLabel(book.cycle.finishedOn)}</dd>
          </dl>
        </section>
        <section className="portal-card">
          <div className="card-heading">
            <h2>Obra y edición</h2>
          </div>
          <dl className="book-facts">
            <dt>Serie</dt>
            <dd>{seriesLabel(book)}</dd>
            <dt>Idioma</dt>
            <dd>{book.edition.language || "Sin especificar"}</dd>
            <dt>Formato</dt>
            <dd>{labelFor(info.formats, book.edition.format)}</dd>
            <dt>ISBN / ID</dt>
            <dd>{book.edition.identifier || "Sin especificar"}</dd>
          </dl>
        </section>
        <section className="portal-card full-width-card">
          <div className="card-heading">
            <h2>Notas personales</h2>
          </div>
          <p className="personal-notes">
            {book.notes || "Todavía no agregaste notas para este libro."}
          </p>
        </section>
      </div>
    </section>
  );
}
