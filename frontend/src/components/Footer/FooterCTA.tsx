import { Link } from "react-router";
import "./FooterCTA.css";
import { useAuth } from "../../context/useAuth";

export default function FooterCTA() {
  const { isAuthenticated } = useAuth();

  if (isAuthenticated) {
    return null;
  }

  return (
    <footer className="footer glass-panel fade-in">
      <div className="footer-content">
        <h2 className="footer-title">Want more? Sign up now!</h2>
        <p className="footer-text">Custom links, powerful analytics, and much more.</p>
      </div>
      <div className="footer-actions">
        <Link to="/login" className="btn-outline">
          Login
        </Link>
        <Link to="/signup" className="btn-primary-glow">
          Sign up
        </Link>
      </div>
    </footer>
  );
}
