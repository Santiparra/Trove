import type {
  Book,
  Dashboard,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type { Choice } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import Icon from "../components/Icon";
import BookCover from "./BookCover";
import StatusSelect from "./StatusSelect";
import { labelFor } from "./presentation";

export default function BookPortal({
  summary,
  statuses,
  busy,
  onList,
  onBook,
  onStatus,
}: {
  summary: Dashboard;
  statuses: Choice[];
  busy: boolean;
  onList: (status: string) => void;
  onBook: (book: Book) => void;
  onStatus: (book: Book, status: string) => void;
}) {
  return (
    <div className="portal-grid">
      <section className="portal-card states-card">
        <div className="card-heading">
          <h2>Tu biblioteca</h2>
          <span>{summary.total} libros</span>
        </div>
        <p className="card-subtitle">Cada historia tiene su momento.</p>
        <div className="state-links">
          {summary.counts.map((item) => (
            <button
              key={item.status}
              onClick={() => onList(item.status)}
              disabled={busy}
              aria-label={`${labelFor(statuses, item.status)}: ${item.count}`}
            >
              <span className={`state-dot dot-${item.status}`} />
              <span>{labelFor(statuses, item.status)}</span>
              <strong>{item.count}</strong>
              <Icon name="chevron" size={14} />
            </button>
          ))}
        </div>
        <button
          className="all-books-link"
          onClick={() => onList("")}
          disabled={busy}
        >
          <Icon name="list" size={16} /> Ver todos los libros{" "}
          <Icon name="chevron" size={15} />
        </button>
      </section>

      <section className="portal-card reading-card">
        <div className="card-heading">
          <h2>Leyendo ahora</h2>
          <button
            className="text-action"
            onClick={() => onList("reading")}
            disabled={busy}
          >
            Ver listado <Icon name="chevron" size={14} />
          </button>
        </div>
        {summary.currentlyReading.length ? (
          <div className="reading-grid">
            {summary.currentlyReading.map((book) => (
              <article className="reading-book" key={book.id}>
                <button
                  className="cover-button"
                  aria-label={`Abrir ${book.title}`}
                  onClick={() => onBook(book)}
                  disabled={busy}
                >
                  <BookCover book={book} compact />
                </button>
                <div>
                  <button
                    className="book-title-button"
                    onClick={() => onBook(book)}
                    disabled={busy}
                  >
                    {book.title}
                  </button>
                  <p>{book.authors.join(" · ") || "Autoría sin especificar"}</p>
                  <StatusSelect
                    book={book}
                    options={statuses}
                    disabled={busy}
                    onChange={onStatus}
                  />
                </div>
              </article>
            ))}
          </div>
        ) : (
          <div className="card-empty">
            <Icon name="books" size={30} />
            <p>No hay lecturas en curso.</p>
            <span>Podés empezar desde tu listado de pendientes.</span>
            <button className="text-action" onClick={() => onList("to-read")}>
              Ver por leer <Icon name="chevron" size={14} />
            </button>
          </div>
        )}
      </section>

      <section className="portal-card distribution-card">
        <div className="card-heading">
          <h2>Tu recorrido</h2>
          <Icon name="chart" size={19} />
        </div>
        <p className="card-subtitle">Una mirada a todos tus libros.</p>
        <div className="distribution-total">
          <strong>{summary.total}</strong>
          <span>libros en tu biblioteca</span>
        </div>
        <div
          className="distribution-bar"
          role="img"
          aria-label={
            summary.total
              ? summary.counts
                  .map(
                    (item) =>
                      `${labelFor(statuses, item.status)}: ${item.count}`,
                  )
                  .join(", ")
              : "Biblioteca vacía"
          }
        >
          {summary.counts
            .filter((item) => item.count > 0)
            .map((item) => (
              <span
                key={item.status}
                className={`segment dot-${item.status}`}
                style={{ flexGrow: item.count }}
                title={`${labelFor(statuses, item.status)}: ${item.count}`}
              />
            ))}
        </div>
        <div className="chart-legend">
          {summary.counts.map((item) => (
            <span key={item.status}>
              <i className={`state-dot dot-${item.status}`} />
              {labelFor(statuses, item.status)}
              <strong>
                {summary.total
                  ? Math.round((item.count / summary.total) * 100)
                  : 0}
                %
              </strong>
            </span>
          ))}
        </div>
      </section>

      <section className="portal-card recent-card">
        <div className="card-heading">
          <h2>Últimos libros leídos</h2>
          <button
            className="text-action"
            onClick={() => onList("read")}
            disabled={busy}
          >
            Ver leídos <Icon name="chevron" size={14} />
          </button>
        </div>
        <p className="card-subtitle">
          Tus cinco últimas finalizaciones con fecha registrada.
        </p>
        {summary.recentlyFinished.length ? (
          <div className="recent-covers">
            {summary.recentlyFinished.map((book) => (
              <button
                key={book.id}
                className="cover-button"
                aria-label={`Abrir ${book.title}`}
                title={book.title}
                onClick={() => onBook(book)}
                disabled={busy}
              >
                <BookCover book={book} />
              </button>
            ))}
          </div>
        ) : (
          <div className="recent-empty">
            <Icon name="books" size={28} />
            <p>Tus próximas historias terminadas aparecerán acá.</p>
            <span>Registrá una fecha de finalización para ordenarlas.</span>
          </div>
        )}
        {summary.readWithoutDate > 0 && (
          <p className="card-footnote">
            {summary.readWithoutDate}{" "}
            {summary.readWithoutDate === 1
              ? "libro leído sin fecha de finalización"
              : "libros leídos sin fecha de finalización"}
            . Siguen en tu listado de leídos.
          </p>
        )}
      </section>
    </div>
  );
}
