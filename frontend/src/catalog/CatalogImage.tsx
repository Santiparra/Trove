import { useEffect, useState, type ReactNode } from "react";
import * as Catalog from "../../bindings/github.com/Santiparra/Trove/internal/desktop/catalog";

export default function CatalogImage({
  url,
  fallback,
}: {
  url: string;
  fallback: ReactNode;
}) {
  const [data, setData] = useState("");
  useEffect(() => {
    let active = true;
    setData("");
    if (url)
      Catalog.CoverDataURL(url)
        .then((value) => {
          if (active) setData(value);
        })
        .catch(() => {});
    return () => {
      active = false;
    };
  }, [url]);
  return data ? (
    <img className="cover-photo" src={data} alt="" />
  ) : (
    <>{fallback}</>
  );
}
