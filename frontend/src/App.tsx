import { useEffect, useState } from "react";
import * as Library from "../bindings/github.com/Santiparra/Trove/internal/desktop/library";
import type { Info } from "../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import type {
  Book,
  Dashboard,
  Input,
  Series,
} from "../bindings/github.com/Santiparra/Trove/internal/books/models";
import Icon, { type IconName } from "./components/Icon";
import BookPortal from "./books/BookPortal";
import BookList from "./books/BookList";
import BookDetails from "./books/BookDetails";
import BookEditor from "./books/BookEditor";
import BookSearch from "./books/BookSearch";
import DeleteDialog from "./books/DeleteDialog";
import { errorMessage, labelFor } from "./books/presentation";
import BackupControl from "./backup/BackupControl";

const media: { key: string; label: string; icon: IconName }[] = [
  { key: "books", label: "Libros", icon: "books" },
  { key: "novels", label: "Novelas ligeras", icon: "novels" },
  { key: "anime", label: "Anime", icon: "anime" },
  { key: "manga", label: "Manga", icon: "manga" },
  { key: "manhwa", label: "Manhwa", icon: "manhwa" },
  { key: "tv", label: "Series de TV", icon: "tv" },
  { key: "movies", label: "Películas", icon: "movies" },
];
type View = "portal" | "list" | "series" | "detail" | "create" | "edit";

export default function App() {
  const [section, setSection] = useState("books");
  const [view, setView] = useState<View>("portal");
  const [returnTo, setReturnTo] = useState<"portal" | "list" | "series">(
    "portal",
  );
  const [info, setInfo] = useState<Info | null>(null);
  const [summary, setSummary] = useState<Dashboard | null>(null);
  const [books, setBooks] = useState<Book[]>([]);
  const [series, setSeries] = useState<Series[]>([]);
  const [query, setQuery] = useState("");
  const [searchText, setSearchText] = useState("");
  const [selectedSeries, setSelectedSeries] = useState<Series | null>(null);
  const [status, setStatus] = useState("");
  const [revision, setRevision] = useState(0);
  const [loading, setLoading] = useState(true);
  const [listLoading, setListLoading] = useState(false);
  const [loadError, setLoadError] = useState("");
  const [listError, setListError] = useState("");
  const [actionError, setActionError] = useState("");
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);
  const [selected, setSelected] = useState<Book | null>(null);
  const [dirty, setDirty] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Book | null>(null);
  const [showStorage, setShowStorage] = useState(false);
  const editing = view === "create" || view === "edit";

  useEffect(() => {
    let active = true;
    setLoading(true);
    Promise.all([Library.Info(), Library.Dashboard(), Library.ListSeries()])
      .then(([appInfo, overview, groups]) => {
        if (!active) return;
        setInfo(appInfo);
        setSummary(overview);
        setSeries(groups);
        setLoadError("");
      })
      .catch((error) => {
        if (active) setLoadError(errorMessage(error));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [revision]);

  useEffect(() => {
    if (view !== "list" && view !== "series") return;
    let active = true;
    setListLoading(true);
    setListError("");
    const timer = setTimeout(() => {
      (view === "series" && selectedSeries
        ? Library.ListSeriesBooks(selectedSeries.id)
        : Library.ListBooks(query, status)
      )
        .then((items) => {
          if (active) {
            setBooks(items);
            setListError("");
          }
        })
        .catch((error) => {
          if (active) {
            setBooks([]);
            setListError(errorMessage(error));
          }
        })
        .finally(() => {
          if (active) setListLoading(false);
        });
    }, 180);
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [view, query, status, revision, selectedSeries]);

  const canLeave = () =>
    !dirty ||
    window.confirm("Tenés cambios sin guardar. ¿Querés descartarlos?");
  function clearMessages() {
    setActionError("");
    setNotice("");
  }
  function navigate(key: string) {
    if (busy || deleteTarget || !canLeave()) return;
    setSection(key);
    setView("portal");
    setSelected(null);
    setQuery("");
    setSearchText("");
    setSelectedSeries(null);
    setStatus("");
    setDirty(false);
    clearMessages();
  }
  function openList(value: string) {
    if (busy || !canLeave()) return;
    setStatus(value);
    setQuery("");
    setSearchText("");
    setSelectedSeries(null);
    setDirty(false);
    setView("list");
    setReturnTo("list");
    setSelected(null);
    clearMessages();
  }
  async function openBook(book: Book) {
    if (busy || !canLeave()) return;
    setBusy(true);
    clearMessages();
    try {
      const current = await Library.GetBook(book.id);
      setSelected(current);
      setReturnTo(
        view === "series" ? "series" : view === "list" ? "list" : "portal",
      );
      setDirty(false);
      setView("detail");
    } catch (error) {
      setActionError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  function openEditor(book: Book | null) {
    if (busy || !canLeave()) return;
    if (!book)
      setReturnTo(
        view === "series" ? "series" : view === "list" ? "list" : "portal",
      );
    setSelected(book);
    setDirty(false);
    setView(book ? "edit" : "create");
    clearMessages();
  }
  function cancelEditor() {
    if (busy || !canLeave()) return;
    setDirty(false);
    setView(view === "edit" ? "detail" : returnTo);
    clearMessages();
  }
  async function save(input: Input) {
    setBusy(true);
    clearMessages();
    try {
      const result = await Library.SaveBook(input);
      setSelected(result);
      setDirty(false);
      setView("detail");
      setRevision((value) => value + 1);
      setNotice("Libro guardado en la biblioteca local.");
    } catch (error) {
      setActionError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  async function changeStatus(book: Book, value: string) {
    if (busy || value === book.cycle.status) return;
    setBusy(true);
    clearMessages();
    try {
      const result = await Library.ChangeBookStatus(book.id, value);
      if (selected?.id === result.id) setSelected(result);
      setRevision((current) => current + 1);
      setNotice("Estado de lectura actualizado.");
    } catch (error) {
      setActionError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  function requestDelete(book: Book) {
    clearMessages();
    setDeleteTarget(book);
  }
  function openSeries(value: Series) {
    if (busy || !canLeave()) return;
    setSelectedSeries(value);
    setSelected(null);
    setView("series");
    setReturnTo("series");
    setDirty(false);
    setQuery("");
    setSearchText("");
    clearMessages();
  }
  function allSearchResults(value: string) {
    if (busy || !canLeave()) return;
    setQuery(value);
    setSearchText(value);
    setStatus("");
    setSelectedSeries(null);
    setView("list");
    setReturnTo("list");
    setDirty(false);
    clearMessages();
  }
  async function updatePersonal(
    operation: () => PromiseLike<Book>,
    message: string,
  ) {
    if (busy) return;
    setBusy(true);
    clearMessages();
    try {
      const result = await operation();
      setSelected((current) => (current?.id === result.id ? result : current));
      setRevision((value) => value + 1);
      setNotice(message);
    } catch (error) {
      setActionError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  async function confirmDelete() {
    if (!deleteTarget || busy) return;
    setBusy(true);
    setActionError("");
    try {
      await Library.DeleteBook(deleteTarget.id);
      if (selected?.id === deleteTarget.id) {
        setSelected(null);
        setView(returnTo);
        setDirty(false);
      }
      setDeleteTarget(null);
      setRevision((value) => value + 1);
      setNotice("Libro eliminado de esta biblioteca.");
    } catch (error) {
      setActionError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  const activeLabel = media.find((item) => item.key === section)?.label;
  const title =
    view === "series"
      ? selectedSeries?.name || "Serie"
      : view === "list"
        ? status && info
          ? labelFor(info.statuses, status)
          : "Todos los libros"
        : "Libros";

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
          {media.map((item) => (
            <button
              key={item.key}
              className={`nav-item ${section === item.key ? "active" : ""}`}
              aria-current={section === item.key ? "page" : undefined}
              onClick={() => navigate(item.key)}
              disabled={busy || !!deleteTarget}
            >
              <Icon name={item.icon} />
              {item.label}
              {item.key === "books" && <span className="nav-dot" />}
            </button>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <span className="connection-dot" />
          Todo queda en tu PC<small>Sin cuentas. Sin proveedores.</small>
        </div>
      </aside>
      <main
        className={`main-area ${section === "books" && view === "portal" ? "portal-main" : ""}`}
      >
        <header className="topbar">
          <span>
            Biblioteca <span className="breadcrumb">/</span>
            <button
              className="breadcrumb-button"
              onClick={() => navigate(section)}
              disabled={busy || !!deleteTarget}
            >
              {activeLabel}
            </button>
            {section === "books" && view !== "portal" && (
              <>
                <span className="breadcrumb">/</span>
                <strong>
                  {editing
                    ? view === "create"
                      ? "Añadir"
                      : "Editar"
                    : view === "detail"
                      ? "Detalle"
                      : title}
                </strong>
              </>
            )}
          </span>
          <div className="topbar-actions">
            <BackupControl
              disabled={busy || !!deleteTarget}
              canOpen={() => {
                if (busy || deleteTarget || !canLeave()) return false;
                if (dirty) {
                  setDirty(false);
                  setView("portal");
                  setSelected(null);
                }
                return true;
              }}
              onRestored={() => {
                setDirty(false);
                setSelected(null);
                setSelectedSeries(null);
                setQuery("");
                setSearchText("");
                setStatus("");
                setSection("books");
                setView("portal");
                setReturnTo("portal");
                setActionError("");
                setNotice("Biblioteca restaurada desde el respaldo.");
                setRevision((value) => value + 1);
              }}
            />
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
          </div>
        </header>
        {showStorage && (
          <div className="storage-panel">
            <strong>Almacenamiento local</strong>
            <code>{info?.dataDir ?? "Cargando…"}</code>
            <span>
              Versión: {info?.version ?? "…"} · Revisión:{" "}
              {info?.revision?.slice(0, 12) ?? "…"} · Build:{" "}
              {info?.buildID ?? "…"} · SQLite · Configuración de resguardo en
              «Respaldo y restauración».
            </span>
          </div>
        )}
        {loadError && (
          <div className="error-banner page-error" role="alert">
            {loadError}
            <button onClick={() => setRevision((value) => value + 1)}>
              Reintentar
            </button>
          </div>
        )}
        {notice && (
          <div className="notice page-notice" role="status">
            {notice}
            <button aria-label="Cerrar aviso" onClick={() => setNotice("")}>
              ×
            </button>
          </div>
        )}
        {actionError && !editing && !deleteTarget && (
          <div className="error-banner page-error" role="alert">
            {actionError}
          </div>
        )}
        {section !== "books" ? (
          <section className="construction">
            <span className="construction-icon">
              <Icon
                name={media.find((item) => item.key === section)!.icon}
                size={52}
              />
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
              Ir a Libros <Icon name="chevron" size={16} />
            </button>
          </section>
        ) : (
          <>
            {(view === "portal" || view === "list" || view === "series") && (
              <>
                {(view === "list" || view === "series") && (
                  <div className="list-navigation">
                    <button
                      className="back-link list-back"
                      onClick={() => navigate("books")}
                    >
                      <Icon name="arrow" size={17} /> Volver al portal
                    </button>
                    {view === "list" && (
                      <nav
                        className="status-tabs"
                        aria-label="Estados de lectura"
                      >
                        <button
                          aria-pressed={status === ""}
                          onClick={() => setStatus("")}
                          disabled={busy}
                        >
                          Todos
                        </button>
                        {info?.statuses.map((option) => (
                          <button
                            key={option.value}
                            aria-pressed={status === option.value}
                            onClick={() => setStatus(option.value)}
                            disabled={busy}
                          >
                            {option.label}
                          </button>
                        ))}
                      </nav>
                    )}
                  </div>
                )}
                <section className="page-heading">
                  <div>
                    <p className="eyebrow">
                      {view === "portal"
                        ? "TU ESPACIO DE LECTURA"
                        : "BIBLIOTECA LOCAL"}
                    </p>
                    <h1>
                      {title}
                      <span className="heading-dot">.</span>
                    </h1>
                    <p>
                      {view === "portal"
                        ? "Tus lecturas, tus pendientes y tu próxima historia."
                        : "Elegí un libro para abrir su ficha o cambiá su estado directamente."}
                    </p>
                  </div>
                  <button
                    className="primary"
                    onClick={() => openEditor(null)}
                    disabled={!info || busy}
                  >
                    <Icon name="plus" size={17} /> Añadir libro
                  </button>
                </section>
                <div className="library-toolbar">
                  <BookSearch
                    value={searchText}
                    disabled={busy}
                    onChange={(value) => {
                      setSearchText(value);
                      if (!value.trim() && view === "list") setQuery("");
                    }}
                    onBook={openBook}
                    onAll={allSearchResults}
                  />
                  <button
                    className="quiet refresh-button"
                    onClick={() => setRevision((value) => value + 1)}
                    disabled={loading || listLoading || busy}
                  >
                    <Icon name="refresh" size={16} /> Actualizar
                  </button>
                  <span className="library-count">
                    {view === "list" || view === "series"
                      ? listLoading
                        ? "Cargando…"
                        : `${books.length} resultados`
                      : loading
                        ? "Actualizando…"
                        : "Vista general"}
                  </span>
                </div>
              </>
            )}
            {view === "portal" && summary && info && (
              <BookPortal
                summary={summary}
                statuses={info.statuses}
                busy={busy || loading}
                onList={openList}
                onBook={openBook}
                onStatus={changeStatus}
                onSeries={openSeries}
              />
            )}
            {view === "portal" && !summary && !loadError && (
              <p className="loading-message" role="status">
                Cargando tu biblioteca…
              </p>
            )}
            {(view === "list" || view === "series") && (
              <>
                {listError && (
                  <div className="error-banner" role="alert">
                    {listError}
                    <button onClick={() => setRevision((value) => value + 1)}>
                      Reintentar
                    </button>
                  </div>
                )}
                {!listError && (
                  <BookList
                    books={books}
                    statuses={info?.statuses ?? []}
                    busy={busy || listLoading}
                    loading={listLoading}
                    onBook={openBook}
                    onStatus={changeStatus}
                    onDelete={requestDelete}
                    onCreate={() => openEditor(null)}
                    onSeries={openSeries}
                  />
                )}
              </>
            )}
            {view === "detail" && selected && info && (
              <BookDetails
                book={selected}
                info={info}
                busy={busy}
                onBack={() => {
                  if (busy || !canLeave()) return;
                  setView(returnTo);
                  setDirty(false);
                  clearMessages();
                }}
                onEdit={() => openEditor(selected)}
                onDelete={requestDelete}
                onStatus={changeStatus}
                onSeries={openSeries}
                onFavorite={(book, value) =>
                  void updatePersonal(
                    () => Library.SetBookFavorite(book.id, value),
                    "Favorito actualizado.",
                  )
                }
                onOwned={(book, value) =>
                  void updatePersonal(
                    () => Library.SetBookOwned(book.id, value),
                    "Propiedad actualizada.",
                  )
                }
                onNotes={(book, notes) =>
                  updatePersonal(
                    () => Library.SaveBookNotes(book.id, notes),
                    "Notas guardadas.",
                  )
                }
                onDirty={setDirty}
              />
            )}
            {editing && info && (
              <BookEditor
                key={view === "edit" ? selected?.id : "new"}
                book={view === "edit" ? selected : null}
                info={info}
                series={series}
                busy={busy}
                error={actionError}
                onCancel={cancelEditor}
                onSave={save}
                onDirty={setDirty}
              />
            )}
            <footer className="page-footer">
              <span>Tu biblioteca es tuya.</span>
              <span>
                Guardado local en SQLite <span className="connection-dot" />
              </span>
            </footer>
          </>
        )}
      </main>
      {deleteTarget && (
        <DeleteDialog
          book={deleteTarget}
          busy={busy}
          error={actionError}
          onCancel={() => {
            if (!busy) {
              setDeleteTarget(null);
              setActionError("");
            }
          }}
          onConfirm={confirmDelete}
        />
      )}
    </div>
  );
}
