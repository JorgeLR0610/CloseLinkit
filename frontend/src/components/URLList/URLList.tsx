import type { URLItem } from "../../types/url";
import "./URLList.css";
import URLListItem from "./URLListItem";

interface Props {
  history: URLItem[];
}

export default function URLList({ history }: Props) {
  return (
    <div className="url-list-container fade-in">
      <h3 className="list-title">List of shortened URLs</h3>
      <div className="url-list">
        {history.map((item) => (
          <URLListItem key={item.shortURL} item={item} />
        ))}
      </div>
    </div>
  );
}
