import type { URLItem } from "../../types/url";
import "./URLList.css";
import URLListItem from "./URLListItem";

interface Props {
  history: URLItem[];
  onDelete?: (shortURL: string) => Promise<void> | void;
}

export default function URLList({ history, onDelete }: Props) {
  return (
    <div className="url-list-container fade-in">
      <h3 className="list-title">List of shortened URLs</h3>
      <div className="url-list">
        {history.map((item) => (
          <URLListItem key={item.shortURL} item={item} onDelete={onDelete} />
        ))}
      </div>
    </div>
  );
}
