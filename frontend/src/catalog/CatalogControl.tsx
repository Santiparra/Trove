import { useState } from "react";
import { Browser } from "@wailsio/runtime";
import * as Catalog from "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog";
import type { CatalogState } from "../../bindings/github.com/Santiparra/Trove/internal/desktop/models";
import { errorMessage } from "../books/presentation";

export default function CatalogControl({
  state,
  disabled,
  onState,
}: {
  state: CatalogState | null;
  disabled: boolean;
  onState: (state: CatalogState) => void;
}) {
  const [open, setOpen] = useState(false);
  const [key, setKey] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  async function save(remove: boolean) {
    setBusy(true);
    setError("");
    try {
      onState(await (remove ? Catalog.ClearKey() : Catalog.SaveKey(key)));
      setKey("");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="catalog-control">
      <button
        className={`profile-tag ${state ? (state.configured ? "keys-configured" : "keys-pending") : "keys-unknown"}`}
        disabled={disabled || busy}
        aria-expanded={open}
        onClick={() => {
          setOpen((value) => !value);
          setKey("");
        }}
      >
        Claves · {state?.configured ? "1" : "0"}
      </button>
      {open && (
        <section
          className="catalog-key-panel"
          aria-label="Claves de proveedores"
        >
          <div className="card-heading">
            <h2>Proveedores</h2>
            <button
              aria-label="Cerrar claves"
              onClick={() => {
                setOpen(false);
                setKey("");
              }}
              disabled={busy}
            >
              ×
            </button>
          </div>
          <strong>
            Hardcover · {state?.configured ? "Configurado" : "Sin clave"}
          </strong>
          <p className="management-hint">
            La clave permite consultar el catálogo de Hardcover desde Trove. Se
            obtiene en tu cuenta de Hardcover; no necesitás una cuenta de Trove.
          </p>
          <button
            className="text-action key-help-link"
            disabled={busy}
            onClick={() => {
              setError("");
              void Browser.OpenURL("https://hardcover.app/account/api").catch(
                () =>
                  setError(
                    "No se pudo abrir Hardcover. Entrá a hardcover.app/account/api desde tu navegador.",
                  ),
              );
            }}
          >
            Obtener clave en Hardcover ↗
          </button>
          <label>
            Clave de catálogo
            <input
              type="password"
              autoComplete="off"
              value={key}
              disabled={busy || disabled}
              onChange={(event) => setKey(event.target.value)}
            />
          </label>
          <p className="management-hint">
            Protegida con Windows en este perfil. El acceso se comprueba al
            buscar.
          </p>
          <div className="catalog-actions">
            <button
              className="secondary"
              disabled={busy || disabled || !key.trim()}
              onClick={() => void save(false)}
            >
              Guardar clave
            </button>
            <button
              className="quiet"
              disabled={busy || disabled || !state?.configured}
              onClick={() => void save(true)}
            >
              Quitar clave
            </button>
          </div>
          <p>Open Library · Sin clave</p>
          <p>RanobeDB · Sin clave</p>
          {error && <p role="alert">{error}</p>}
        </section>
      )}
    </div>
  );
}
