import './Header.css'

export default function Header() {
  return (
    <header className="header">
      <div className="logo-container">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="logo-icon">
          <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path>
          <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>
        </svg>
        <span className="logo-text">CloseLinkit</span>
      </div>
      <nav className="nav-links">
        <button className="btn-text">Prices</button>
        <button className="btn-text">Login</button>
        <button className="btn-primary">Sign up</button>
      </nav>
    </header>
  )
}
