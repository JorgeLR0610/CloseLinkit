import CopyButton from './CopyButton'
import './RecentURLBox.css'

export default function RecentURLBox({ shortURL }: { shortURL: string }) {

  return (
    <div className="recent-box glass-panel fade-in">
      <div className="recent-top">
        <span className="recent-url">{shortURL}</span>
        <CopyButton textToCopy={shortURL} className='copy-btn'/>
      </div>
      <div className="recent-bottom">
        <div className="social-icons">
          <div className="social-icon whatsapp" title="WhatsApp">W</div>
          <div className="social-icon facebook" title="Facebook">F</div>
          <div className="social-icon x" title="X">X</div>
        </div>
      </div>
    </div>
  )
}
