import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Choice } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import Icon from "../components/Icon";
import StatusSelect from "./StatusSelect";

export default function BookList({
  books,
  statuses,
  busy,
  loading,
  onBook,
  onStatus,
  onDelete,
  onCreate,
}: {
  books: Book[];
  statuses: Choice[];
  busy: boolean;
  loading: boolean;
  onBook: (book: Book) => void;
  onStatus: (book: Book, status: string) => void;
  onDelete: (book: Book) => void;
  onCreate: () => void;
}) {
  return (
    <section
      className="book-list"
      aria-label="Biblioteca de libros"
      aria-busy={loading}
    >
      {!loading && !books.length ? (
        <div className="empty-state">
          <span className="empty-book">
            <Icon name="books" size={32} />
          </span>
          <h2>No hay libros en este listado</h2>
          <p>Probá otro estado o búsqueda, o añadí un libro manualmente.</p>
          <button className="secondary" onClick={onCreate}>
            <Icon name="plus" size={16} /> Añadir libro
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
                className="book-open"
                aria-label={`Abrir ${book.title}`}
                onClick={() => onBook(book)}
                disabled={busy}
              >
                <span className="book-spine" aria-hidden="true">
                  {book.title.slice(0, 1).toUpperCase()}
                </span>
                <span className="book-text">
                  <strong>{book.title}</strong>
                  <span>
                    {book.authors.join(" · ") || "Autoría sin especificar"}
                  </span>
                  {book.series && (
                    <small>
                      {book.series.name}
                      {book.series.position !== null
                        ? ` · #${book.series.position}`
                        : ""}
                    </small>
                  )}
                </span>
              </button>
              <StatusSelect
                book={book}
                options={statuses}
                disabled={busy}
                onChange={onStatus}
              />
              <button
                className="icon-button danger-icon"
                title="Eliminar libro"
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
