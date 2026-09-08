import "./Header.css";
import logo from "../../assets/logo.svg";

export default function Header() {
  return (
    <header className="header">
      <div className="logo-container">
        <img src={logo} alt="CloseLinkit" className="header-logo" />
        <span className="logo-text">CloseLinkit</span>
      </div>
      <div className="header-actions">
        <button type="button" className="btn login-btn">
          Log in
        </button>
        <button type="button" className="btn signup-btn">
          Sign Up
        </button>
      </div>
    </header>
  );
}
