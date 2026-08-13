import "./Header.css";
import logo from "../../assets/logo.svg";

export default function Header() {
  return (
    <header className="header">
      <div className="logo-container">
        <img src={logo} alt="CloseLinkit" className="header-logo" />
        <span className="logo-text">CloseLinkit</span>
      </div>
    </header>
  );
}
