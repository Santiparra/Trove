import { Fragment, useLayoutEffect, useRef, useState } from "react";
import { Browser, Clipboard } from "@wailsio/runtime";

function webLinks(text: string) {
  const links: { start: number; url: string }[] = [];
  for (const match of text.matchAll(/https?:\/\/[^\s<>"']+/g)) {
    let value = match[0].replace(/[.,;!?\]}]+$/g, "");
    while (
      value.endsWith(")") &&
      (value.match(/\)/g) || []).length > (value.match(/\(/g) || []).length
    )
      value = value.slice(0, -1);
    try {
      const url = new URL(value);
      if (
        !url.username &&
        !url.password &&
        ["http:", "https:"].includes(url.protocol)
      )
        links.push({ start: match.index!, url: value });
    } catch {}
  }
  return links;
}
export default function DescriptionText({ text }: { text: string }) {
  const [expanded, setExpanded] = useState(false);
  const [overflow, setOverflow] = useState(false);
  const [copied, setCopied] = useState("");
  const [error, setError] = useState("");
  const body = useRef<HTMLParagraphElement>(null);
  useLayoutEffect(() => {
    setExpanded(false);
    setCopied("");
    setError("");
    const measure = () => {
      if (!body.current) return;
      const line = parseFloat(getComputedStyle(body.current).lineHeight) || 24;
      setOverflow(body.current.scrollHeight > Math.ceil(line * 7) + 1);
    };
    measure();
    void document.fonts?.ready.then(measure);
    const observer =
      typeof ResizeObserver !== "undefined"
        ? new ResizeObserver(measure)
        : null;
    if (body.current) observer?.observe(body.current);
    window.addEventListener("resize", measure);
    return () => {
      observer?.disconnect();
      window.removeEventListener("resize", measure);
    };
  }, [text]);
  const links = webLinks(text);
  let offset = 0;
  const content = links.map((link) => {
    const before = text.slice(offset, link.start);
    offset = link.start + link.url.length;
    return (
      <Fragment key={link.start}>
        {before}
        <a
          className="source-link"
          href={link.url}
          target="_blank"
          rel="noopener noreferrer"
          onClick={(event) => {
            event.preventDefault();
            setError("");
            void Browser.OpenURL(link.url).catch(() =>
              setError("No se pudo abrir el enlace."),
            );
          }}
        >
          {link.url}
        </a>
      </Fragment>
    );
  });
  const unique = [...new Set(links.map((link) => link.url))];
  return (
    <div className="profile-description">
      <p ref={body} className={!expanded ? "description-collapsed" : ""}>
        {content}
        {text.slice(offset)}
      </p>
      {overflow && (
        <button
          className="text-action"
          aria-expanded={expanded}
          onClick={() => setExpanded((value) => !value)}
        >
          {expanded ? "Leer menos" : "Leer más"}
        </button>
      )}
      {!!unique.length && (
        <div className="description-links" aria-label="Fuentes enlazadas">
          {unique.map((url) => (
            <div className="description-source" key={url}>
              <span title={url}>{new URL(url).hostname}</span>
              <button
                className="text-action"
                onClick={() => {
                  setError("");
                  void Clipboard.SetText(url)
                    .then(() => setCopied(url))
                    .catch(() => setError("No se pudo copiar el enlace."));
                }}
                aria-label={`Copiar enlace ${url}`}
              >
                {copied === url ? "Copiado" : "Copiar enlace"}
              </button>
            </div>
          ))}
        </div>
      )}
      {error && <small role="alert">{error}</small>}
    </div>
  );
}
