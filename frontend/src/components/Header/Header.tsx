import { Link, useNavigate } from "react-router";
import "./Header.css";
import logo from "../../assets/logo.svg";
import { useAuth } from "../../context/useAuth";
import toast from "react-hot-toast";

export default function Header() {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await logout();
      toast.success("Logged out successfully");
      navigate("/");
    } catch {
      toast.error("Logout failed. Please try again.");
    }
  };

  return (
    <header className="header">
      <div className="logo-container">
        <Link to="/" className="logo-link">
          <img src={logo} alt="CloseLinkit" className="header-logo" />
          <span className="logo-text">CloseLinkit</span>
        </Link>
      </div>
      <nav className="header-nav">
        {isAuthenticated ? (
          <>
            {user && (
              <span className="user-badge" title={user.email}>
                {user.email}
              </span>
            )}
            <button onClick={handleLogout} className="header-btn-outline" aria-label="Logout">
              Logout
            </button>
          </>
        ) : (
          <>
            <Link to="/login" className="header-btn-outline">
              Login
            </Link>
            <Link to="/signup" className="header-btn-primary">
              Sign up
            </Link>
          </>
        )}
      </nav>
    </header>
  );
}
