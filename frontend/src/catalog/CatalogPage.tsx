import { useEffect, useState, type FormEvent } from "react";
import * as Catalog from "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog";
import type {
  CatalogPreview,
  CatalogState,
} from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import type {
  Work,
  Edition,
} from "../../bindings/github.com/Santiparra/Trove/internal/catalog/models";
import type {
  Book,
  CatalogResult,
} from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import Icon from "../components/Icon";
import { errorMessage } from "../books/presentation";
import CatalogImage from "./CatalogImage";

export default function CatalogPage({
  target,
  onBack,
  onManual,
  onDone,
  onBusy,
}: {
  target: Book | null;
  onBack: () => void;
  onManual: () => void;
  onDone: (book: Book, message: string) => void;
  onBusy: (value: boolean) => void;
}) {
  const [state, setState] = useState<CatalogState | null>(null);
  const [key, setKey] = useState("");
  const [query, setQuery] = useState(target?.title || "");
  const [works, setWorks] = useState<Work[]>([]);
  const [work, setWork] = useState<Work | null>(null);
  const [editions, setEditions] = useState<Edition[]>([]);
  const [preview, setPreview] = useState<CatalogPreview | null>(null);
  const [series, setSeries] = useState("");
  const [review, setReview] = useState<CatalogResult | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [searched, setSearched] = useState(false);
  useEffect(() => {
    let active = true;
    Catalog.State()
      .then((value) => {
        if (active) setState(value);
      })
      .catch((reason) => {
        if (active) setError(errorMessage(reason));
      });
    return () => {
      active = false;
    };
  }, []);
  async function action(operation: () => PromiseLike<unknown>) {
    if (busy) return;
    setBusy(true);
    onBusy(true);
    setError("");
    setNotice("");
    try {
      await operation();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
      onBusy(false);
    }
  }
  async function search(event: FormEvent) {
    event.preventDefault();
    await action(async () => {
      const result = await Catalog.Search(query);
      setWorks(result.works);
      setSearched(true);
      setWork(null);
      setEditions([]);
      setPreview(null);
      setReview(null);
      setNotice(result.warnings?.join(" ") || "");
    });
  }
  async function resolve(id: string) {
    const value = await Catalog.Preview(id);
    setPreview(value);
    setSeries("");
    setReview(null);
  }
  async function select(candidate: Work) {
    await action(async () => {
      setWork(candidate);
      setPreview(null);
      setReview(null);
      if (candidate.matchedEditionId) {
        await resolve(candidate.matchedEditionId);
      } else {
        setEditions(await Catalog.Editions(candidate.id));
      }
    });
  }
  const snapshot = preview?.snapshot;
  const effectiveTarget = target || review?.book || null;
  return (
    <section className="catalog-page" aria-label="Catálogo de Libros">
      <button className="back-link" disabled={busy} onClick={onBack}>
        <Icon name="arrow" size={17} /> Volver
      </button>
      <header className="page-heading">
        <div>
          <p className="eyebrow">HARDCOVER · EDICIONES · DATOS LOCALES</p>
          <h1>
            {target ? "Vincular edición" : "Catálogo de Libros"}
            <span className="heading-dot">.</span>
          </h1>
          <p>
            Buscá una edición o seguí con un registro manual. Consultar no
            agrega nada a tu biblioteca.
          </p>
        </div>
        <button className="secondary" disabled={busy} onClick={onManual}>
          {target ? "Seguir con edición manual" : "Crear manualmente"}
        </button>
      </header>
      <details className="catalog-settings" open={!state?.configured}>
        <summary>
          Acceso al catálogo ·{" "}
          {state?.configured
            ? "clave guardada en este perfil"
            : "sin clave configurada"}
        </summary>
        <p>
          Clave de lectura de catálogo Hardcover. Se protege con tu cuenta
          Windows y no entra en los respaldos. No se consulta ni sincroniza tu
          biblioteca de Hardcover.
        </p>
        <label>
          Clave de catálogo
          <input
            type="password"
            autoComplete="off"
            value={key}
            disabled={busy}
            onChange={(event) => setKey(event.target.value)}
            placeholder="Pegá el token de lectura, no tu contraseña"
          />
        </label>
        <div className="catalog-actions">
          <button
            className="secondary"
            disabled={busy || !key.trim()}
            onClick={() =>
              void action(async () => {
                setState(await Catalog.SaveKey(key));
                setKey("");
                setNotice(
                  "Clave protegida guardada. La búsqueda comprobará su acceso al catálogo.",
                );
              })
            }
          >
            Guardar clave
          </button>
          <button
            className="quiet"
            disabled={busy || !state?.configured}
            onClick={() =>
              void action(async () => {
                setState(await Catalog.ClearKey());
                setKey("");
                setNotice(
                  "Clave quitada. Tu biblioteca y snapshots siguen disponibles.",
                );
              })
            }
          >
            Quitar clave
          </button>
        </div>
      </details>
      <form className="catalog-query" onSubmit={(event) => void search(event)}>
        <label>
          <span className="sr-only">Buscar en Hardcover</span>
          <input
            type="search"
            value={query}
            maxLength={300}
            disabled={busy}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Título, autor, serie o ISBN"
          />
        </label>
        <button
          className="primary"
          disabled={busy || !state?.configured || query.trim().length < 2}
        >
          <Icon name="search" size={17} /> Buscar catálogo
        </button>
      </form>
      {busy && (
        <p className="catalog-working" role="status">
          Consultando o guardando…
        </p>
      )}
      {error && (
        <div className="error-banner" role="alert">
          {error}
        </div>
      )}
      {notice && (
        <div className="notice" role="status">
          {notice}
        </div>
      )}
      {!preview && (
        <>
          {!work && (
            <div
              className="catalog-results"
              aria-label="Candidatos de catálogo"
            >
              {works.map((candidate) => (
                <button
                  className="catalog-work"
                  key={candidate.id}
                  disabled={busy}
                  onClick={() => void select(candidate)}
                >
                  <span className="catalog-mini-cover">
                    <CatalogImage
                      url={candidate.coverURL}
                      fallback={<Icon name="books" size={28} />}
                    />
                  </span>
                  <span>
                    <strong>{candidate.title}</strong>
                    <small>
                      {candidate.authors?.join(" · ") ||
                        "Revisar autoría en el detalle"}
                    </small>
                    <small>
                      Hardcover · obra {candidate.id}
                      {candidate.matchedEditionId ? " · coincidencia ISBN" : ""}
                    </small>
                  </span>
                  <Icon name="chevron" size={17} />
                </button>
              ))}
            </div>
          )}
          {searched && !work && !works.length && (
            <div className="empty-state">
              <h2>No hay candidatos en el catálogo principal</h2>
              <p>
                Probá el ISBN o una variante de título/autor. Podés crear la
                edición manual y vincularla después.
              </p>
              <button className="secondary" onClick={onManual}>
                Crear manualmente
              </button>
            </div>
          )}
          {work && (
            <section className="catalog-editions">
              <div className="card-heading">
                <h2>{work.title}</h2>
                <button
                  className="text-action"
                  disabled={busy}
                  onClick={() => {
                    setWork(null);
                    setEditions([]);
                  }}
                >
                  Volver a candidatos
                </button>
              </div>
              <p>
                Se muestran hasta 20 ediciones usadas. Para una edición
                concreta, buscá su ISBN.
              </p>
              {!editions.length && (
                <p>
                  No se encontraron ediciones seleccionables. Podés usar el alta
                  manual.
                </p>
              )}
              <div className="edition-grid">
                {editions.map((edition) => (
                  <button
                    className="edition-choice"
                    key={edition.id}
                    disabled={busy}
                    onClick={() => void action(() => resolve(edition.id))}
                  >
                    <strong>{edition.title}</strong>
                    <span>
                      {edition.language || "Idioma no informado"} ·{" "}
                      {edition.format || "Formato no informado"}
                    </span>
                    <span>
                      {edition.publisher || "Editorial no informada"} ·{" "}
                      {edition.pages ?? "?"} páginas
                    </span>
                    <small>
                      {edition.declared
                        ?.filter((id) => id.scheme.startsWith("isbn"))
                        .map((id) => `${id.scheme}: ${id.value}`)
                        .join(" · ") || "Sin ISBN declarado"}{" "}
                      · edición {edition.id}
                    </small>
                  </button>
                ))}
              </div>
            </section>
          )}
        </>
      )}
      {snapshot && preview && (
        <section className="catalog-selection" aria-label="Revisión de edición">
          <div className="catalog-preview-hero">
            <div className="catalog-preview-cover">
              <CatalogImage
                url={snapshot.coverURL}
                fallback={<Icon name="books" size={40} />}
              />
            </div>
            <div>
              <p className="eyebrow">EDICIÓN SELECCIONADA</p>
              <h2>{snapshot.edition.title}</h2>
              <p>
                {snapshot.work.authors?.join(" · ") || "Autoría no informada"}
              </p>
              <p>
                {snapshot.edition.language || "Idioma no informado"} ·{" "}
                {snapshot.edition.format || "Formato no informado"} ·{" "}
                {snapshot.edition.publisher || "Editorial no informada"}
              </p>
              <p>
                {snapshot.edition.pages ?? "?"} páginas ·{" "}
                {snapshot.edition.publishedOn || "Publicación sin fecha"}
              </p>
              <div className="catalog-identifiers">
                {snapshot.identifiers?.map((id) => (
                  <span key={`${id.scheme}:${id.value}`}>
                    {id.scheme}: {id.value}
                    {id.derived ? " (equivalente)" : ""}
                  </span>
                ))}
              </div>
              <small>
                Hardcover: obra {snapshot.work.id} · edición{" "}
                {snapshot.edition.id}
              </small>
              {snapshot.coverURL && (
                <p>
                  Fuente de portada:{" "}
                  {snapshot.sources?.cover?.provider || snapshot.provider}
                </p>
              )}
            </div>
          </div>
          {snapshot.warnings?.length > 0 && (
            <ul className="catalog-warnings">
              {snapshot.warnings.map((message, index) => (
                <li key={index}>{message}</li>
              ))}
            </ul>
          )}
          {snapshot.description && (
            <section className="catalog-description">
              <h3>
                Descripción{" "}
                <small>
                  Fuente:{" "}
                  {snapshot.sources?.description?.provider || snapshot.provider}
                </small>
              </h3>
              <p>{snapshot.description}</p>
            </section>
          )}
          {!effectiveTarget && (
            <label className="catalog-series">
              Serie local al agregar
              <select
                value={series}
                disabled={busy}
                onChange={(event) => setSeries(event.target.value)}
              >
                <option value="">
                  Sin asignar serie; puedo editarla después
                </option>
                {snapshot.work.series?.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name}
                    {item.position !== null ? ` · #${item.position}` : ""}
                  </option>
                ))}
              </select>
            </label>
          )}
          {effectiveTarget && (
            <div className="catalog-local-comparison">
              <h3>Registro local: {effectiveTarget.title}</h3>
              <p>
                {effectiveTarget.authors?.join(" · ") || "Sin autoría"} ·
                ISBN/UID:{" "}
                {effectiveTarget.edition.identifier || "Sin identificador"}
              </p>
              <p>
                {effectiveTarget.edition.language || "Idioma sin informar"} ·{" "}
                {effectiveTarget.edition.format || "Formato sin informar"}
              </p>
              <p>
                Vincular conserva título, autoría, serie, edición local, notas,
                estados y fechas. Revisá que corresponda a la misma edición.
              </p>
            </div>
          )}
          {review && (
            <div className="notice" role="status">
              {review.message}
            </div>
          )}
          <div className="catalog-actions">
            <button
              className="quiet"
              disabled={busy}
              onClick={() =>
                void action(async () => {
                  setReview(null);
                  setPreview(null);
                  setEditions(await Catalog.Editions(snapshot.work.id));
                  setWork(snapshot.work);
                })
              }
            >
              Elegir otra edición
            </button>
            {effectiveTarget ? (
              <button
                className="primary"
                disabled={busy}
                onClick={() =>
                  void action(async () => {
                    const linked = await Catalog.Link(
                      effectiveTarget.id,
                      preview.token,
                    );
                    onDone(
                      linked,
                      "Edición vinculada. Se conservaron tus datos personales.",
                    );
                  })
                }
              >
                Vincular con registro local
              </button>
            ) : (
              <button
                className="primary"
                disabled={busy}
                onClick={() =>
                  void action(async () => {
                    const result = await Catalog.Add(preview.token, series);
                    if (result.kind === "review") {
                      setReview(result);
                    } else {
                      onDone(result.book, result.message);
                    }
                  })
                }
              >
                Agregar edición a Libros
              </button>
            )}
            {review && (
              <button
                className="secondary"
                disabled={busy}
                onClick={() => onDone(review.book, review.message)}
              >
                Abrir registro existente
              </button>
            )}
          </div>
        </section>
      )}
    </section>
  );
}
