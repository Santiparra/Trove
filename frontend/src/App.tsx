import { FormEvent, useEffect, useState } from "react";
import * as Library from "../bindings/github.com/Santiparra/Trove/internal/desktop/library";
import type { Info } from "../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import {
  Input,
  type Book,
  type Series,
} from "../bindings/github.com/Santiparra/Trove/internal/books/models";

const media = [
  ["books", "Libros", "▤"],
  ["novels", "Novelas ligeras", "▧"],
  ["anime", "Anime", "▷"],
  ["manga", "Manga", "▥"],
  ["manhwa", "Manhwa", "▦"],
  ["tv", "Series de TV", "▣"],
  ["movies", "Películas", "◈"],
] as const;

type Draft = {
  id: string;
  title: string;
  authors: string;
  seriesName: string;
  seriesPosition: string;
  language: string;
  format: string;
  identifier: string;
  status: string;
  notes: string;
  addedOn: string;
  startedOn: string;
  finishedOn: string;
};

function emptyDraft(): Draft {
  return {
    id: "",
    title: "",
    authors: "",
    seriesName: "",
    seriesPosition: "",
    language: "",
    format: "",
    identifier: "",
    status: "to-read",
    notes: "",
    addedOn: "",
    startedOn: "",
    finishedOn: "",
  };
}

function bookDraft(book: Book): Draft {
  return {
    id: book.id,
    title: book.title,
    authors: book.authors.join("\n"),
    seriesName: book.series?.name ?? "",
    seriesPosition: book.series?.position?.toString() ?? "",
    language: book.edition.language,
    format: book.edition.format,
    identifier: book.edition.identifier,
    status: book.cycle.status,
    notes: book.notes,
    addedOn: book.addedOn,
    startedOn: book.cycle.startedOn,
    finishedOn: book.cycle.finishedOn,
  };
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function dateLabel(value: string): string {
  if (!value) return "Sin registrar";
  const [year, month, day] = value.split("-");
  return `${day}/${month}/${year}`;
}

export default function App() {
  const [section, setSection] = useState("books");
  const [info, setInfo] = useState<Info | null>(null);
  const [books, setBooks] = useState<Book[]>([]);
  const [series, setSeries] = useState<Series[]>([]);
  const [query, setQuery] = useState("");
  const [refresh, setRefresh] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [selected, setSelected] = useState<Book | null>(null);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [initialDraft, setInitialDraft] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [notice, setNotice] = useState("");
  const [showStorage, setShowStorage] = useState(false);

  useEffect(() => {
    let active = true;
    Library.Info()
      .then((value) => {
        if (active) setInfo(value);
      })
      .catch((error) => {
        if (active) setLoadError(errorMessage(error));
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    let active = true;
    const timer = setTimeout(() => {
      setLoading(true);
      Promise.all([Library.ListBooks(query), Library.ListSeries()])
        .then(([items, groups]) => {
          if (!active) return;
          setBooks(items);
          setSeries(groups);
          setLoadError("");
        })
        .catch((error) => {
          if (active) setLoadError(errorMessage(error));
        })
        .finally(() => {
          if (active) setLoading(false);
        });
    }, 180);
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [query, refresh]);

  const dirty = draft !== null && JSON.stringify(draft) !== initialDraft;
  const canLeave = () =>
    !dirty ||
    window.confirm("Tenés cambios sin guardar. ¿Querés descartarlos?");

  function closeEditor() {
    if (saving || !canLeave()) return;
    setDraft(null);
    setSaveError("");
  }

  function navigate(key: string) {
    if (saving || !canLeave()) return;
    setSection(key);
    setDraft(null);
    setSelected(null);
    setSaveError("");
    setNotice("");
  }

  function openEditor(book?: Book) {
    if (saving || !canLeave()) return;
    const value = book ? bookDraft(book) : emptyDraft();
    setDraft(value);
    setInitialDraft(JSON.stringify(value));
    setSaveError("");
    setNotice("");
    setSelected(book ?? null);
  }

  function selectBook(book: Book) {
    if (saving || !canLeave()) return;
    setSelected(book);
    setDraft(null);
    setSaveError("");
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    if (!draft || saving) return;
    setSaving(true);
    setSaveError("");
    setNotice("");
    try {
      const book = await Library.SaveBook(
        new Input({
          ...draft,
          authors: draft.authors.split("\n"),
          seriesPosition:
            draft.seriesPosition === "" ? null : Number(draft.seriesPosition),
        }),
      );
      setSelected(book);
      setDraft(null);
      setInitialDraft("");
      setQuery("");
      setRefresh((value) => value + 1);
      setNotice("Libro guardado en la biblioteca local.");
    } catch (error) {
      setSaveError(errorMessage(error));
    } finally {
      setSaving(false);
    }
  }

  const statusLabel = (value: string) =>
    info?.statuses.find((item) => item.value === value)?.label ?? value;
  const formatLabel = (value: string) =>
    info?.formats.find((item) => item.value === value)?.label ?? value;
  const change = (field: keyof Draft, value: string) =>
    setDraft((current) => (current ? { ...current, [field]: value } : null));
  const activeLabel = media.find(([key]) => key === section)?.[1];

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-mark">t.</span>
          <div>
            Trove<small>TU UNIVERSO, EN ORDEN</small>
          </div>
        </div>
        <p className="nav-caption">BIBLIOTECA</p>
        <nav aria-label="Medios">
          {media.map(([key, label, icon]) => (
            <button
              key={key}
              className={`nav-item ${section === key ? "active" : ""}`}
              aria-current={section === key ? "page" : undefined}
              onClick={() => navigate(key)}
              disabled={saving}
            >
              <span aria-hidden="true" className="nav-icon">
                {icon}
              </span>
              {label}
              {key === "books" && <span className="nav-dot" />}
            </button>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <span className="connection-dot" />
          Todo queda en tu PC<small>Sin cuentas. Sin proveedores.</small>
        </div>
      </aside>

      <main className="main-area">
        <header className="topbar">
          <span>
            Biblioteca <span className="breadcrumb">/</span>{" "}
            <strong>{activeLabel}</strong>
          </span>
          <button
            className="profile-tag"
            onClick={() => setShowStorage(!showStorage)}
            aria-expanded={showStorage}
          >
            <span className="connection-dot" />
            {info?.profile === "habitual"
              ? "Uso habitual"
              : info?.profile === "test"
                ? "PRUEBA · no publicado"
                : "DESARROLLO"}
            <span aria-hidden="true"> ⓘ</span>
          </button>
        </header>
        {showStorage && (
          <div className="storage-panel">
            <strong>Almacenamiento local</strong>
            <code>{info?.dataDir ?? "Cargando…"}</code>
            <span>
              Build: {info?.buildID ?? "…"} · SQLite · Sin respaldo automático
              en esta entrega.
            </span>
          </div>
        )}

        {section !== "books" ? (
          <section className="construction">
            <span className="construction-icon" aria-hidden="true">
              ◇
            </span>
            <p className="eyebrow">PRÓXIMAMENTE EN TROVE</p>
            <h1>{activeLabel}</h1>
            <h2>En construcción</h2>
            <p>
              Este espacio ya tiene su lugar.
              <br />
              Por ahora, podés empezar a construir tu biblioteca de libros.
            </p>
            <button className="primary" onClick={() => navigate("books")}>
              Ir a Libros <span aria-hidden="true">→</span>
            </button>
          </section>
        ) : (
          <>
            <section className="page-heading">
              <div>
                <p className="eyebrow">TUS HISTORIAS, TU REGISTRO</p>
                <h1>
                  Libros<span className="heading-dot">.</span>
                </h1>
                <p>Un lugar para lo que leíste y todo lo que viene.</p>
              </div>
              <button
                className="primary"
                onClick={() => openEditor()}
                disabled={!info || saving}
              >
                <span aria-hidden="true">＋</span> Añadir libro
              </button>
            </section>

            <div className="library-toolbar">
              <label className="search-box">
                <span aria-hidden="true">⌕</span>
                <input
                  type="search"
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder="Buscar por título en tu biblioteca…"
                  aria-label="Buscar por título"
                />
              </label>
              <button
                className="quiet"
                onClick={() => setRefresh((value) => value + 1)}
                disabled={loading}
                aria-label="Actualizar listado"
              >
                ↻ Actualizar
              </button>
              <span className="library-count">
                {loading
                  ? "Cargando…"
                  : `${books.length} ${query ? "resultados" : books.length === 1 ? "libro" : "libros"}`}
              </span>
            </div>
            {notice && (
              <div className="notice" role="status">
                ✓ {notice}
                <button aria-label="Cerrar aviso" onClick={() => setNotice("")}>
                  ×
                </button>
              </div>
            )}
            {loadError && (
              <div className="error-banner" role="alert">
                {loadError}
                <button onClick={() => setRefresh((value) => value + 1)}>
                  Reintentar
                </button>
              </div>
            )}

            <div
              className={`library-layout ${draft || selected ? "with-panel" : ""}`}
            >
              <section
                className="book-list"
                aria-label="Biblioteca de libros"
                aria-busy={loading}
              >
                {books.length === 0 && !loading && !loadError ? (
                  <div className="empty-state">
                    <span className="empty-book" aria-hidden="true">
                      ▤
                    </span>
                    <h2>
                      {query
                        ? "No encontramos ese título"
                        : "Tu próxima historia empieza acá"}
                    </h2>
                    <p>
                      {query
                        ? "Probá otra búsqueda o añadí un libro manualmente."
                        : "Añadí tu primer libro. No necesitás ISBN, conexión ni una ficha completa."}
                    </p>
                    <button
                      className="secondary"
                      onClick={() => openEditor()}
                      disabled={!info}
                    >
                      ＋ Añadir libro manualmente
                    </button>
                  </div>
                ) : (
                  <>
                    <div className="list-heading">
                      <span>OBRA Y AUTORÍA</span>
                      <span>LECTURA</span>
                    </div>
                    {books.map((book) => (
                      <button
                        className={`book-row ${selected?.id === book.id ? "selected" : ""}`}
                        key={book.id}
                        onClick={() => selectBook(book)}
                        disabled={saving}
                      >
                        <span className="book-spine" aria-hidden="true">
                          {book.title.slice(0, 1).toUpperCase()}
                        </span>
                        <span className="book-text">
                          <strong>{book.title}</strong>
                          <span>
                            {book.authors.join(" · ") ||
                              "Autoría sin especificar"}
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
                        <span className={`status status-${book.cycle.status}`}>
                          {statusLabel(book.cycle.status)}
                        </span>
                        <span className="row-chevron" aria-hidden="true">
                          ›
                        </span>
                      </button>
                    ))}
                  </>
                )}
              </section>

              {draft ? (
                <aside
                  className="detail-panel"
                  aria-label={draft.id ? "Editar libro" : "Añadir libro"}
                >
                  <div className="panel-heading">
                    <div>
                      <p className="eyebrow">REGISTRO MANUAL</p>
                      <h2>{draft.id ? "Editar libro" : "Añadir libro"}</h2>
                    </div>
                    <button
                      className="icon-button"
                      onClick={closeEditor}
                      aria-label="Cerrar formulario"
                      disabled={saving}
                    >
                      ×
                    </button>
                  </div>
                  <form onSubmit={save}>
                    <fieldset disabled={saving}>
                      <legend>Identificación</legend>
                      <label>
                        Título <span className="required">*</span>
                        <input
                          autoFocus
                          required
                          value={draft.title}
                          onChange={(e) => change("title", e.target.value)}
                          placeholder="Título del libro"
                        />
                      </label>
                      <label>
                        Autor o autores
                        <textarea
                          rows={2}
                          value={draft.authors}
                          onChange={(e) => change("authors", e.target.value)}
                          placeholder="Un nombre por línea"
                        />
                        <small>
                          Un autor por línea. Podés completarlo después.
                        </small>
                      </label>
                      <div className="form-row series-fields">
                        <label>
                          Serie
                          <input
                            list="series-options"
                            value={draft.seriesName}
                            onChange={(e) =>
                              change("seriesName", e.target.value)
                            }
                            placeholder="Sin serie"
                          />
                          <datalist id="series-options">
                            {series.map((group) => (
                              <option key={group.id} value={group.name} />
                            ))}
                          </datalist>
                        </label>
                        <label>
                          N.º de orden
                          <input
                            type="number"
                            min="0"
                            step="any"
                            value={draft.seriesPosition}
                            onChange={(e) =>
                              change("seriesPosition", e.target.value)
                            }
                            placeholder="Ej. 1"
                          />
                        </label>
                      </div>
                      <small>
                        Elegí una serie existente o escribí un nombre para
                        vincular los libros.
                      </small>
                    </fieldset>
                    <fieldset disabled={saving}>
                      <legend>
                        Edición <span>OPCIONAL</span>
                      </legend>
                      <div className="form-row">
                        <label>
                          Idioma
                          <input
                            value={draft.language}
                            onChange={(e) => change("language", e.target.value)}
                            placeholder="Ej. Inglés"
                          />
                        </label>
                        <label>
                          Formato
                          <select
                            value={draft.format}
                            onChange={(e) => change("format", e.target.value)}
                          >
                            {info?.formats.map((option) => (
                              <option key={option.value} value={option.value}>
                                {option.label}
                              </option>
                            ))}
                          </select>
                        </label>
                      </div>
                      <label>
                        ISBN / identificador
                        <input
                          value={draft.identifier}
                          onChange={(e) => change("identifier", e.target.value)}
                          placeholder="No es necesario para guardar"
                        />
                      </label>
                    </fieldset>
                    <fieldset disabled={saving}>
                      <legend>Tu lectura</legend>
                      <label>
                        Estado
                        <select
                          value={draft.status}
                          onChange={(e) => change("status", e.target.value)}
                        >
                          {info?.statuses.map((option) => (
                            <option key={option.value} value={option.value}>
                              {option.label}
                            </option>
                          ))}
                        </select>
                      </label>
                      <div className="form-row">
                        <label>
                          Inicio
                          <input
                            type="date"
                            value={draft.startedOn}
                            onChange={(e) =>
                              change("startedOn", e.target.value)
                            }
                          />
                        </label>
                        <label>
                          Finalización
                          <input
                            type="date"
                            value={draft.finishedOn}
                            onChange={(e) =>
                              change("finishedOn", e.target.value)
                            }
                          />
                        </label>
                      </div>
                      <label>
                        Añadido el
                        <input
                          type="date"
                          value={draft.addedOn}
                          onChange={(e) => change("addedOn", e.target.value)}
                        />
                        <small>
                          {draft.id
                            ? "Fecha de incorporación a tu biblioteca."
                            : "Si lo dejás vacío, se registra la fecha de hoy."}
                        </small>
                      </label>
                      <label>
                        Notas personales
                        <textarea
                          rows={4}
                          value={draft.notes}
                          onChange={(e) => change("notes", e.target.value)}
                          placeholder="Tus impresiones, pendientes, ideas…"
                        />
                      </label>
                    </fieldset>
                    {saveError && (
                      <div className="error-banner" role="alert">
                        {saveError}
                      </div>
                    )}
                    <div className="form-actions">
                      <button
                        type="button"
                        className="quiet"
                        onClick={closeEditor}
                        disabled={saving}
                      >
                        Cancelar
                      </button>
                      <button
                        type="submit"
                        className="primary"
                        disabled={saving}
                      >
                        {saving ? "Guardando…" : "Guardar libro"}
                      </button>
                    </div>
                  </form>
                </aside>
              ) : selected ? (
                <aside className="detail-panel" aria-label="Detalle del libro">
                  <div className="panel-heading">
                    <p className="eyebrow">EN TU BIBLIOTECA</p>
                    <button
                      className="icon-button"
                      aria-label="Cerrar detalle"
                      onClick={() => setSelected(null)}
                    >
                      ×
                    </button>
                  </div>
                  <div className="detail-cover" aria-hidden="true">
                    {selected.title.slice(0, 1).toUpperCase()}
                    <span>TROVE / LIBROS</span>
                  </div>
                  <h2 className="detail-title">{selected.title}</h2>
                  <p className="detail-author">
                    {selected.authors.join(" · ") || "Autoría sin especificar"}
                  </p>
                  <span className={`status status-${selected.cycle.status}`}>
                    {statusLabel(selected.cycle.status)}
                  </span>
                  <button
                    className="secondary edit-button"
                    onClick={() => openEditor(selected)}
                  >
                    Editar libro
                  </button>
                  <dl className="book-facts">
                    <dt>Serie</dt>
                    <dd>
                      {selected.series ? (
                        <span>
                          {selected.series.name}
                          {selected.series.position !== null
                            ? ` · #${selected.series.position}`
                            : ""}
                        </span>
                      ) : (
                        "Sin serie"
                      )}
                    </dd>
                    <dt>Idioma</dt>
                    <dd>{selected.edition.language || "Sin especificar"}</dd>
                    <dt>Formato</dt>
                    <dd>{formatLabel(selected.edition.format)}</dd>
                    <dt>ISBN / ID</dt>
                    <dd>{selected.edition.identifier || "Sin especificar"}</dd>
                    <dt>Añadido</dt>
                    <dd>{dateLabel(selected.addedOn)}</dd>
                    <dt>Inicio</dt>
                    <dd>{dateLabel(selected.cycle.startedOn)}</dd>
                    <dt>Finalización</dt>
                    <dd>{dateLabel(selected.cycle.finishedOn)}</dd>
                  </dl>
                  <div className="detail-notes">
                    <h3>Notas personales</h3>
                    <p>
                      {selected.notes ||
                        "Todavía no agregaste notas para este libro."}
                    </p>
                  </div>
                  <p className="local-hint">
                    Registro local · Sin vinculación a proveedores
                  </p>
                </aside>
              ) : null}
            </div>
            <footer className="page-footer">
              <span>Tu biblioteca es tuya.</span>
              <span>
                Guardado local en SQLite <span className="connection-dot" />
              </span>
            </footer>
          </>
        )}
      </main>
    </div>
  );
}
