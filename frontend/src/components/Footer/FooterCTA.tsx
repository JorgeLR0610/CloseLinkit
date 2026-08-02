import './FooterCTA.css'

export default function FooterCTA() {
  return (
    <footer className="footer glass-panel fade-in">
      <div className="footer-content">
        <h2 className="footer-title">Want more? Sign up now!</h2>
        <p className="footer-text">Custom links, powerful analytics, and much more.</p>
      </div>
      <div className="footer-actions">
        <button className="btn-outline">Login</button>
        <button className="btn-primary-glow">Sign up</button>
      </div>
    </footer>
  )
}
