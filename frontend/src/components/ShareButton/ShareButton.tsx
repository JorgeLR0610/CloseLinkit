import type { IconType } from "react-icons";

interface ShareButtonsProps {
  href: string;
  icon: IconType;
  label: string;
  className?: string;
}

export default function ShareButton({ href, icon: Icon, label, className }: ShareButtonsProps) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className={`share-button  ${className ?? ""}`}
      aria-label={`Share on ${label}`}
    >
      <Icon />
    </a>
  );
}
