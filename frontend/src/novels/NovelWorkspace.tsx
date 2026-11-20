import { useEffect, useState } from "react";
import * as Novels from "../../bindings/github.com/Santiparra/Trove/internal/desktop/novels";
import type {
  Novel,
  Volume,
  Input,
} from "../../bindings/github.com/Santiparra/Trove/internal/novels/models";
import type { Book } from "../../bindings/github.com/Santiparra/Trove/internal/books/models";
import type {
  Info,
  NovelPreview,
} from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import BookProfile from "../books/BookProfile";
import { dateLabel, errorMessage } from "../books/presentation";
import NovelEditor from "./NovelEditor";
import VolumeEditor from "./VolumeEditor";
import Icon from "../components/Icon";
import type {
  Candidate,
  SearchResult,
} from "../../bindings/github.com/Santiparra/Trove/internal/novelcatalog/models";
import NovelSearch from "./NovelSearch";
import CatalogNovel, { PublicationSources } from "./CatalogNovel";
import { representedNovel } from "./presentation";

function progress(n: Novel) {
  return `${n.completed ?? "?"}/${n.total ?? "?"}`;
}
function nextNumber(n: Novel) {
  return n.completed == null ||
    n.status === "read" ||
    (n.total != null && n.completed >= n.total)
    ? 0
    : n.completed + 1;
}
function asInput(n: Novel): Input {
  return {
    id: n.id,
    title: n.title,
    authors: n.authors,
    status: n.status,
    total: n.total,
    completed: n.completed,
    notes: n.notes,
    favorite: n.favorite,
    addedOn: n.addedOn,
  };
}

export default function NovelWorkspace({
  info,
  onBusy,
  onDirty,
  canLeave,
  onNotice,
  localOnly = false,
}: {
  info: Info;
  onBusy: (value: boolean) => void;
  onDirty: (value: boolean) => void;
  canLeave: () => boolean;
  onNotice: (message: string) => void;
  localOnly?: boolean;
}) {
  const [items, setItems] = useState<Novel[]>([]);
  const [selected, setSelected] = useState<Novel | null>(null);
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [revision, setRevision] = useState(0);
  const [editing, setEditing] = useState(false);
  const [volumeEdit, setVolumeEdit] = useState<{
    volume: Volume | null;
    number: number;
  } | null>(null);
  const [notes, setNotes] = useState("");
  const [all, setAll] = useState(false);
  const [searchText, setSearchText] = useState("");
  const [authorID, setAuthorID] = useState("");
  const [remote, setRemote] = useState<SearchResult | null>(null);
  const [remoteLoading, setRemoteLoading] = useState(false);
  const [remoteError, setRemoteError] = useState("");
  const [page, setPage] = useState(1);
  const [preview, setPreview] = useState<NovelPreview | null>(null);
  const [linkTarget, setLinkTarget] = useState<Novel | null>(null);
  const [linking, setLinking] = useState(false);
  const [linkQuery, setLinkQuery] = useState("");
  const [viewedNumber, setViewedNumber] = useState<number | null>(null);
  const [metadataError, setMetadataError] = useState("");
  const [metadataLoading, setMetadataLoading] = useState(false);
  const shownNumber = selected
    ? (viewedNumber ??
      (nextNumber(selected) ||
        selected.completed ||
        selected.volumes[selected.volumes.length - 1]?.number ||
        0))
    : 0;
  useEffect(() => {
    let active = true;
    setRemote(null);
    setRemoteError("");
    setRemoteLoading(false);
    if (
      selected ||
      preview ||
      editing ||
      volumeEdit ||
      localOnly ||
      query.trim().length < 2 ||
      status
    )
      return;
    setRemoteLoading(true);
    const timer = setTimeout(() => {
      Novels.SearchCatalog(query, authorID, page, "list")
        .then((value) => {
          if (active) setRemote(value);
        })
        .catch((reason) => {
          if (active) setRemoteError(errorMessage(reason));
        })
        .finally(() => {
          if (active) setRemoteLoading(false);
        });
    }, 350);
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [
    query,
    authorID,
    page,
    status,
    selected?.id,
    !!preview,
    editing,
    volumeEdit,
    localOnly,
    revision,
  ]);
  useEffect(() => {
    let active = true;
    setMetadataError("");
    setMetadataLoading(false);
    if (
      !selected?.catalog ||
      !shownNumber ||
      selected.volumes.some((v) => v.number === shownNumber && v.catalog) ||
      preview ||
      editing ||
      volumeEdit ||
      localOnly
    )
      return;
    setMetadataLoading(true);
    Novels.RefreshCatalog(selected.id, shownNumber, false)
      .then((value) => {
        if (active) setSelected(value);
      })
      .catch((reason) => {
        if (active) setMetadataError(errorMessage(reason));
      })
      .finally(() => {
        if (active) setMetadataLoading(false);
      });
    return () => {
      active = false;
    };
  }, [
    selected?.id,
    selected?.catalog?.id,
    shownNumber,
    !!preview,
    editing,
    volumeEdit,
    localOnly,
  ]);
  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    const timer = setTimeout(
      () => {
        Novels.List(query, status)
          .then((value) => {
            if (active) setItems(value);
          })
          .catch((reason) => {
            if (active) setError(errorMessage(reason));
          })
          .finally(() => {
            if (active) setLoading(false);
          });
      },
      query ? 140 : 0,
    );
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [query, status, revision]);
  useEffect(
    () => setNotes(selected?.notes || ""),
    [selected?.id, selected?.notes],
  );
  useEffect(() => {
    if (!editing && !volumeEdit)
      onDirty(!!selected && notes !== selected.notes);
  }, [notes, selected?.notes, selected?.id, editing, volumeEdit, onDirty]);
  function leave(action: () => void) {
    if (busy || !canLeave()) return;
    onDirty(false);
    setNotes(selected?.notes || "");
    setError("");
    action();
  }
  async function action(operation: () => PromiseLike<void>) {
    if (busy) return;
    setBusy(true);
    onBusy(true);
    setError("");
    try {
      await operation();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
      onBusy(false);
    }
  }
  function commit(n: Novel, message: string, showCurrent = false) {
    setSelected(n);
    if (showCurrent) setViewedNumber(null);
    setEditing(false);
    setVolumeEdit(null);
    setRevision((value) => value + 1);
    onDirty(!editing && !volumeEdit && notes !== n.notes);
    onNotice(message);
  }
  function open(n: Novel) {
    leave(
      () =>
        void action(async () => {
          setSelected(await Novels.Get(n.id));
          setViewedNumber(null);
          setPreview(null);
          setLinking(false);
        }),
    );
  }
  function search(value: string) {
    leave(() => {
      setAuthorID(
        selected?.catalog?.authors.find((author) => author.name === value)
          ?.id || "",
      );
      setSelected(null);
      setPreview(null);
      setLinking(false);
      setViewedNumber(null);
      setQuery(value);
      setSearchText(value);
      setPage(1);
      setStatus("");
      setAll(true);
    });
  }
  function allResults(value: string) {
    leave(() => {
      setQuery(value);
      setSearchText(value);
      setPage(1);
      setAll(true);
      setStatus("");
    });
  }
  function candidate(item: Candidate, target: Novel | null = null) {
    leave(
      () =>
        void action(async () => {
          const value = await Novels.OpenCatalog(
            item.id,
            target ? nextNumber(target) || target.completed || 1 : 1,
            item.matchedBookID || "",
            item.matchedReleaseID || "",
          );
          if (
            value.match.kind === "existing" &&
            value.match.novel &&
            !value.token
          ) {
            setSelected(value.match.novel);
            setViewedNumber(
              item.matchedBookID
                ? value.series.books.find(
                    (ref) => ref.id === item.matchedBookID,
                  )?.number || null
                : null,
            );
            setPreview(null);
            setLinking(false);
            return;
          }
          setLinkTarget(
            target ||
              (value.match.kind === "existing" ? value.match.novel : null),
          );
          setPreview(value);
          setLinking(false);
        }),
    );
  }
  function choosePublication() {
    if (!selected) return;
    leave(
      () =>
        void action(async () => {
          setLinkTarget(selected);
          setPreview(
            await Novels.PrepareCatalog(selected.id, shownNumber || 1),
          );
        }),
    );
  }
  if (preview)
    return (
      <CatalogNovel
        key={preview.token}
        initial={preview}
        target={linkTarget}
        info={info}
        localOnly={localOnly}
        onBusy={onBusy}
        onSearch={search}
        onBack={() => setPreview(null)}
        onDone={(n, message, number) => {
          setSelected(n);
          setPreview(null);
          setLinkTarget(null);
          setLinking(false);
          setViewedNumber(number || null);
          onDirty(false);
          setRevision((value) => value + 1);
          if (message) onNotice(message);
        }}
      />
    );
  if (editing)
    return (
      <NovelEditor
        key={selected?.id || "new"}
        novel={selected}
        info={info}
        busy={busy}
        error={error}
        onDirty={onDirty}
        onCancel={() => leave(() => setEditing(false))}
        onSave={(input) =>
          void action(async () =>
            commit(
              await Novels.Save(input),
              "Novela guardada en la biblioteca local.",
            ),
          )
        }
      />
    );
  if (volumeEdit && selected)
    return (
      <VolumeEditor
        key={volumeEdit.volume?.id || `new-${volumeEdit.number}`}
        volume={volumeEdit.volume}
        number={volumeEdit.number}
        total={selected.total}
        info={info}
        busy={busy}
        error={error}
        onDirty={onDirty}
        onCancel={() => leave(() => setVolumeEdit(null))}
        onSave={(input) =>
          void action(async () =>
            commit(
              await Novels.SaveVolume(selected.id, input),
              "Datos del volumen guardados.",
            ),
          )
        }
      />
    );
  if (selected) {
    const next = nextNumber(selected);
    const shown = shownNumber;
    const current = selected.volumes.find((volume) => volume.number === shown);
    const registeredNumbers = new Set(
      selected.volumes.map((volume) => volume.number),
    );
    let numberToRegister = next || 1;
    while (registeredNumbers.has(numberToRegister)) numberToRegister++;
    if (selected.total != null && numberToRegister > selected.total) {
      numberToRegister = 1;
      while (registeredNumbers.has(numberToRegister)) numberToRegister++;
    }
    const book: Book = {
      id: selected.id,
      medium: "light-novels",
      title: selected.title,
      authors: selected.authors.length
        ? selected.authors
        : selected.catalog?.authors.map((author) => author.name) || [],
      series: null,
      edition: current?.edition || {
        id: "",
        language: "",
        format: "",
        identifier: "",
        owned: false,
        identifiers: [],
        catalog: null,
      },
      cycle: current?.cycle || {
        id: "",
        status: selected.status,
        startedOn: "",
        finishedOn: "",
      },
      notes: selected.notes,
      addedOn: selected.addedOn,
      favorite: selected.favorite,
    };
    return (
      <BookProfile
        book={book}
        ariaLabel="Detalle de novela"
        subtitle={`${progress(selected)} volúmenes completados${shown ? ` · Volumen ${shown}${current?.title || current?.catalog?.title ? `: ${current?.title || current?.catalog?.title}` : ""}` : ""}`}
        description={
          current?.description ||
          current?.catalog?.description ||
          selected.catalog?.description
        }
        coverURL={current?.catalog?.coverURL}
        editionInfo={current?.catalog || undefined}
        formatLabel={
          current?.catalog?.releaseFormat === "print" &&
          current?.edition.format === "other"
            ? "Impreso"
            : undefined
        }
        sourceDetails={
          current?.catalog ? (
            <PublicationSources publication={current.catalog} />
          ) : undefined
        }
        busy={busy}
        onBack={() =>
          leave(() => {
            setSelected(null);
            setViewedNumber(null);
            setLinking(false);
          })
        }
        onSearch={search}
        management={
          <>
            <select
              aria-label={`Estado de ${selected.title}`}
              value={selected.status}
              disabled={busy}
              onChange={(event) =>
                void action(async () =>
                  commit(
                    await Novels.ChangeStatus(selected.id, event.target.value),
                    "Estado de la novela actualizado.",
                  ),
                )
              }
            >
              {info.statuses.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.value === "read" ? "Leída" : option.label}
                </option>
              ))}
            </select>
            <strong className="novel-progress">{progress(selected)}</strong>
            <span className="management-hint">Volúmenes completados</span>
            {viewedNumber != null && !!next && viewedNumber !== next && (
              <button
                className="quiet"
                disabled={busy}
                onClick={() => setViewedNumber(null)}
              >
                Volver al volumen actual {next}
              </button>
            )}
            {selected.total != null &&
              selected.completed === selected.total &&
              selected.status !== "read" && (
                <button
                  className="primary"
                  disabled={busy}
                  onClick={() =>
                    void action(async () =>
                      commit(
                        await Novels.ChangeStatus(selected.id, "read"),
                        "Novela marcada como leída.",
                      ),
                    )
                  }
                >
                  Marcar novela como leída
                </button>
              )}
            <button
              className="primary"
              disabled={
                busy ||
                !next ||
                (selected.status === "reading" &&
                  current?.cycle.status === "reading")
              }
              onClick={() =>
                void action(async () =>
                  commit(
                    await Novels.StartVolume(selected.id),
                    `Volumen ${next} iniciado.`,
                    true,
                  ),
                )
              }
            >
              Empezar volumen{next ? ` ${next}` : ""}
            </button>
            <button
              className="secondary"
              disabled={busy || !next}
              onClick={() =>
                void action(async () =>
                  commit(
                    await Novels.FinishVolume(selected.id),
                    "Volumen terminado. Progreso actualizado.",
                    true,
                  ),
                )
              }
            >
              {next && selected.total === next
                ? "Terminar último volumen y marcar leída"
                : "Volumen terminado"}
            </button>
            <button
              className="quiet"
              disabled={busy}
              onClick={() => leave(() => setEditing(true))}
            >
              Editar novela y progreso
            </button>
            <button
              className={`flag-button ${selected.favorite ? "is-active" : ""}`}
              aria-pressed={selected.favorite}
              disabled={busy}
              onClick={() =>
                void action(async () =>
                  commit(
                    await Novels.Save({
                      ...asInput(selected),
                      favorite: !selected.favorite,
                    }),
                    "Favorita actualizada.",
                  ),
                )
              }
            >
              Favorita
            </button>
            <button
              className="quiet danger-text"
              disabled={busy}
              onClick={() => {
                if (
                  !canLeave() ||
                  !window.confirm(
                    `¿Eliminar «${selected.title}» y todos sus volúmenes?`,
                  )
                )
                  return;
                void action(async () => {
                  await Novels.Delete(selected.id);
                  setSelected(null);
                  setRevision((value) => value + 1);
                  onDirty(false);
                  onNotice("Novela y volúmenes eliminados.");
                });
              }}
            >
              Eliminar novela
            </button>
            <small>Alta: {dateLabel(selected.addedOn)}</small>
            {shown > 0 && (
              <button
                className="secondary"
                disabled={busy || metadataLoading}
                onClick={() =>
                  leave(() =>
                    setVolumeEdit({ volume: current || null, number: shown }),
                  )
                }
              >
                Editar datos del volumen {shown}
              </button>
            )}
            {selected.catalog ? (
              <>
                <small className="management-hint">
                  RanobeDB · Novela ligera
                </small>
                <button
                  className="quiet"
                  disabled={busy || localOnly}
                  onClick={choosePublication}
                >
                  Elegir volumen o publicación
                </button>
                <button
                  className="quiet"
                  disabled={busy || localOnly || metadataLoading}
                  onClick={() =>
                    void action(async () => {
                      const updated = await Novels.RefreshCatalog(
                        selected.id,
                        shown || 0,
                        true,
                      );
                      setMetadataError("");
                      commit(updated, "Metadata de novela actualizada.");
                    })
                  }
                >
                  Refrescar catálogo
                </button>
                <button
                  className="quiet"
                  disabled={busy}
                  onClick={() => {
                    if (
                      window.confirm(
                        "¿Quitar el vínculo y sus snapshots? Se conservan progreso, fechas y datos personales.",
                      )
                    )
                      void action(async () =>
                        commit(
                          await Novels.UnlinkCatalog(selected.id),
                          "Catálogo desvinculado.",
                        ),
                      );
                  }}
                >
                  Desvincular catálogo
                </button>
              </>
            ) : (
              <button
                className="quiet"
                disabled={busy || localOnly}
                onClick={() =>
                  leave(() => {
                    setLinkQuery(selected.title);
                    setLinking((value) => !value);
                  })
                }
              >
                Vincular catálogo
              </button>
            )}
            {metadataLoading && (
              <small role="status">Consultando metadata del volumen…</small>
            )}
            {metadataError && (
              <p className="management-hint" role="alert">
                {metadataError}
              </p>
            )}
            {error && (
              <p className="error-banner" role="alert">
                {error}
              </p>
            )}
          </>
        }
      >
        {linking && (
          <section className="edition-picker" aria-label="Vincular novela">
            <div className="card-heading">
              <h2>Buscar novela para vincular</h2>
              <button className="quiet" onClick={() => setLinking(false)}>
                Cerrar búsqueda
              </button>
            </div>
            <NovelSearch
              value={linkQuery}
              onChange={setLinkQuery}
              disabled={busy}
              localOnly={localOnly}
              catalogOnly
              onAll={() => {}}
              onNovel={() => {}}
              onCandidate={(item) => candidate(item, selected)}
            />
          </section>
        )}
        {!!current?.catalog?.warnings.length && (
          <ul className="catalog-warnings">
            {current.catalog.warnings.map((warning, index) => (
              <li key={index}>{warning}</li>
            ))}
          </ul>
        )}
        <section className="novel-volumes" aria-label="Volúmenes de la novela">
          <div className="card-heading">
            <h2>Volúmenes</h2>
            <button
              className="secondary"
              disabled={
                busy ||
                (selected.total != null &&
                  selected.volumes.length >= selected.total)
              }
              onClick={() =>
                leave(() =>
                  setVolumeEdit({
                    volume: null,
                    number: numberToRegister,
                  }),
                )
              }
            >
              Registrar volumen
            </button>
          </div>
          {!selected.volumes.length && (
            <p className="management-hint">
              Sin volúmenes registrados. Empezar, terminar o declarar un
              progreso conocido crea sus registros; una cantidad desconocida no
              inventa volúmenes.
            </p>
          )}
          <div className="novel-volume-grid">
            {selected.volumes.map((volume) => (
              <button
                className={`novel-volume ${volume.number === next ? "is-current" : ""}`}
                key={volume.id}
                aria-label={`Abrir volumen ${volume.number}`}
                disabled={busy}
                onClick={() => leave(() => setViewedNumber(volume.number))}
              >
                <strong>
                  Volumen {volume.number}
                  {volume.title || volume.catalog?.title
                    ? ` · ${volume.title || volume.catalog?.title}`
                    : ""}
                </strong>
                <span>
                  {
                    info.statuses.find(
                      (option) => option.value === volume.cycle.status,
                    )?.label
                  }
                  {volume.edition.owned ? " · En propiedad" : ""}
                  {volume.favorite ? " · Favorito" : ""}
                </span>
                <small>
                  {volume.cycle.startedOn
                    ? `Inicio: ${dateLabel(volume.cycle.startedOn)}`
                    : "Inicio desconocido"}{" "}
                  ·{" "}
                  {volume.cycle.finishedOn
                    ? `Fin: ${dateLabel(volume.cycle.finishedOn)}`
                    : "Fin desconocido"}
                </small>
              </button>
            ))}
          </div>
        </section>
        <section className="portal-card profile-notes">
          <div className="card-heading">
            <h2>Notas de la novela</h2>
            <button
              className="small-save"
              disabled={busy || notes === selected.notes}
              onClick={() =>
                void action(async () =>
                  commit(
                    await Novels.Save({ ...asInput(selected), notes }),
                    "Notas de la novela guardadas.",
                  ),
                )
              }
            >
              Guardar notas
            </button>
          </div>
          <label>
            <span className="sr-only">Notas de la novela</span>
            <textarea
              rows={4}
              disabled={busy}
              value={notes}
              onChange={(event) => setNotes(event.target.value)}
            />
          </label>
        </section>
      </BookProfile>
    );
  }
  const visible =
    all || query || status
      ? items
      : items.filter(
          (novel) => novel.status === "reading" || novel.status === "up-next",
        );
  return (
    <section className="novel-workspace" aria-label="Biblioteca de novelas">
      <div className="page-heading">
        <h1>{all || query || status ? "Novelas ligeras" : "Tus novelas"}</h1>
        <button
          className="primary"
          disabled={busy}
          onClick={() =>
            leave(() => {
              setSelected(null);
              setEditing(true);
            })
          }
        >
          <Icon name="plus" size={17} /> Añadir novela
        </button>
      </div>
      <div className="library-toolbar">
        <NovelSearch
          value={searchText}
          onChange={(value) => {
            setSearchText(value);
            setAuthorID("");
          }}
          authorID={authorID}
          disabled={busy}
          localOnly={localOnly}
          onNovel={open}
          onCandidate={(item) => candidate(item)}
          onAll={allResults}
        />
        <button
          className="secondary"
          disabled={busy}
          onClick={() => {
            setAll(true);
            setStatus("");
            setQuery("");
            setSearchText("");
            setAuthorID("");
            setPage(1);
          }}
        >
          Ver todas las novelas
        </button>
      </div>
      <div className="status-tabs" aria-label="Estados de novelas">
        <button
          className={!status ? "active" : ""}
          disabled={busy}
          onClick={() => {
            setStatus("");
            setAll(true);
          }}
        >
          Todas
        </button>
        {info.statuses.map((option) => (
          <button
            className={status === option.value ? "active" : ""}
            disabled={busy}
            key={option.value}
            onClick={() => {
              setStatus(option.value);
              setAll(true);
            }}
          >
            {option.value === "read" ? "Leídas" : option.label}
          </button>
        ))}
      </div>
      {error && (
        <p className="error-banner" role="alert">
          {error}
          <button onClick={() => setRevision((value) => value + 1)}>
            Reintentar
          </button>
        </p>
      )}
      {loading && <p role="status">Cargando novelas…</p>}
      {!loading && !items.length && !error && (
        <div className="empty-state">
          <h2>
            {query || status
              ? "No hay coincidencias locales"
              : "Tu biblioteca de novelas está vacía"}
          </h2>
          {!query && !status && (
            <p>Agregá una novela y registrá sus volúmenes y progreso.</p>
          )}
        </div>
      )}
      {!loading && !!items.length && !visible.length && (
        <p className="management-hint">
          No hay novelas en curso. Podés consultar todas o elegir un estado.
        </p>
      )}
      <div className="novel-library-grid">
        {visible.map((novel) => (
          <button
            className="novel-library-card"
            disabled={busy}
            key={novel.id}
            aria-label={`Abrir novela ${novel.title}`}
            onClick={() => open(novel)}
          >
            <Icon name="books" size={28} />
            <div>
              <strong>{novel.title}</strong>
              <span>
                {novel.authors.join(" · ") || "Autoría sin especificar"}
              </span>
              <span>
                {
                  info.statuses.find((option) => option.value === novel.status)
                    ?.label
                }
              </span>
            </div>
            <div className="novel-card-progress">
              <strong>{progress(novel)}</strong>
              <small>
                {nextNumber(novel)
                  ? `Volumen ${nextNumber(novel)}`
                  : novel.status === "read"
                    ? "Lectura completada"
                    : "Progreso por indicar"}
              </small>
            </div>
          </button>
        ))}
      </div>
      {query && !status && !localOnly && (
        <section className="catalog-results" aria-label="Resultados RanobeDB">
          {remoteLoading && <p role="status">Buscando en RanobeDB…</p>}
          {remoteError && (
            <p className="error-banner" role="alert">
              {remoteError}
              <button onClick={() => setRevision((value) => value + 1)}>
                Reintentar catálogo
              </button>
            </p>
          )}
          {remote?.items
            .filter((item) => !representedNovel(item, items))
            .map((item) => (
              <button
                className="catalog-work"
                key={item.id}
                disabled={busy}
                onClick={() => candidate(item)}
              >
                <span>
                  <strong>{item.title}</strong>
                  <small>{item.originalTitle}</small>
                </span>
                <Icon name="chevron" size={17} />
              </button>
            ))}
          {remote && !remote.items.length && (
            <p className="management-hint">Sin coincidencias en RanobeDB.</p>
          )}
          {!!remote && remote.pages > 1 && (
            <div className="catalog-actions">
              <button
                className="quiet"
                disabled={busy || page <= 1}
                onClick={() => setPage((value) => value - 1)}
              >
                Anterior
              </button>
              <span>
                Página {page} de {remote.pages}
              </span>
              <button
                className="quiet"
                disabled={busy || page >= remote.pages}
                onClick={() => setPage((value) => value + 1)}
              >
                Siguiente
              </button>
            </div>
          )}
        </section>
      )}
    </section>
  );
}
