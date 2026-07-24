import type { URLItem } from '../App'
import './URLList.css'

export default function URLList({ history }: { history: URLItem[] }) {
  const copyToClipboard = (url: string) => {
    navigator.clipboard.writeText(url)
  }

  return (
    <div className="url-list-container fade-in">
      <h3 className="list-title">List of shortened URLs</h3>
      <div className="url-list">
        {history.map((item, index) => (
          <div key={index} className="url-item glass-panel">
            <span className="original-url" title={item.original}>{item.original}</span>
            <div className="short-url-group">
              <span className="short-url">{item.short}</span>
              <button onClick={() => copyToClipboard(item.short)} className="copy-btn-small">Copy</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
