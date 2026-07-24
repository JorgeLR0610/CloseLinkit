import './RecentURLBox.css'

export default function RecentURLBox({ shortURL }: { shortURL: string }) {
  const copyToClipboard = () => {
    navigator.clipboard.writeText(shortURL)
  }

  return (
    <div className="recent-box glass-panel fade-in">
      <div className="recent-top">
        <span className="recent-url">{shortURL}</span>
        <button onClick={copyToClipboard} className="copy-btn">Copy</button>
      </div>
      <div className="recent-bottom">
        <span className="share-text">Links to share on popular apps using their logos</span>
        <div className="social-icons">
          <div className="social-icon whatsapp" title="WhatsApp">W</div>
          <div className="social-icon facebook" title="Facebook">F</div>
          <div className="social-icon x" title="X">X</div>
        </div>
      </div>
    </div>
  )
}
