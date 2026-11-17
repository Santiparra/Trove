import type {
  Book,
  Series,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Choice } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import Icon from "../components/Icon";
import StatusSelect from "./StatusSelect";
import { readingLabels } from "./readingMedia";

export default function BookList({
  medium,
  books,
  statuses,
  busy,
  loading,
  onBook,
  onStatus,
  onDelete,
  onCreate,
  onSeries,
}: {
  medium: string;
  books: Book[];
  statuses: Choice[];
  busy: boolean;
  loading: boolean;
  onBook: (book: Book) => void;
  onStatus: (book: Book, status: string) => void;
  onDelete: (book: Book) => void;
  onCreate: () => void;
  onSeries: (series: Series) => void;
}) {
  const labels = readingLabels(medium);
  return (
    <section
      className="book-list"
      aria-label={labels.library}
      aria-busy={loading}
    >
      {!loading && !books.length ? (
        <div className="empty-state">
          <span className="empty-book">
            <Icon name="books" size={32} />
          </span>
          <h2>No hay {labels.plural} en este listado</h2>
          <p>
            Probá otro estado o búsqueda, o añadí un {labels.item} manualmente.
          </p>
          <button className="secondary" onClick={onCreate}>
            <Icon name="plus" size={16} /> {labels.add}
          </button>
        </div>
      ) : (
        <>
          <div className="list-heading">
            <span>OBRA Y AUTORÍA</span>
            <span>ESTADO DE LECTURA / ACCIONES</span>
          </div>
          {books.map((book) => (
            <article className="book-row" key={book.id}>
              <button
                className="book-cover-link"
                aria-label={`Ver ficha de ${book.title}`}
                onClick={() => onBook(book)}
                disabled={busy}
              >
                <span className="book-spine" aria-hidden="true">
                  {book.title.slice(0, 1).toUpperCase()}
                </span>
              </button>
              <div className="book-text">
                {book.series && (
                  <button
                    className="series-title"
                    onClick={() => onSeries(book.series!)}
                    disabled={busy}
                  >
                    {book.series.name}
                    {book.series.position !== null
                      ? ` · #${book.series.position}`
                      : ""}
                  </button>
                )}
                <button
                  className="book-title-button list-book-title"
                  aria-label={`Abrir ${book.title}`}
                  onClick={() => onBook(book)}
                  disabled={busy}
                >
                  {book.title}
                </button>
                <span>
                  {book.authors.join(" · ") || "Autoría sin especificar"}
                </span>
              </div>
              <StatusSelect
                book={book}
                options={statuses}
                disabled={busy}
                onChange={onStatus}
              />
              <button
                className="icon-button danger-icon"
                title={`Eliminar ${labels.item}`}
                aria-label={`Eliminar ${book.title}`}
                onClick={() => onDelete(book)}
                disabled={busy}
              >
                <Icon name="trash" size={18} />
              </button>
            </article>
          ))}
        </>
      )}
    </section>
  );
}
