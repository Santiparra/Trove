import { useEffect, useRef, useState, type ReactNode } from "react";
import * as Catalog from "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog";

export default function CatalogImage({
  url,
  fallback,
  lazy = false,
}: {
  url: string;
  fallback: ReactNode;
  lazy?: boolean;
}) {
  const [data, setData] = useState("");
  const [visible, setVisible] = useState(
    !lazy || typeof IntersectionObserver === "undefined",
  );
  const frame = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (visible || !frame.current) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          setVisible(true);
          observer.disconnect();
        }
      },
      { rootMargin: "200px" },
    );
    observer.observe(frame.current);
    return () => observer.disconnect();
  }, [visible]);
  useEffect(() => {
    let active = true;
    setData("");
    if (url && visible)
      Catalog.CoverDataURL(url)
        .then((value) => {
          if (active) setData(value);
        })
        .catch(() => {});
    return () => {
      active = false;
    };
  }, [url, visible]);
  const image = data ? (
    <img className="cover-photo" src={data} alt="" />
  ) : (
    <>{fallback}</>
  );
  return lazy ? (
    <div className="lazy-cover" ref={frame}>
      {image}
    </div>
  ) : (
    image
  );
}
