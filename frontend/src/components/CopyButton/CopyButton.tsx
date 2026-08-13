import { useState } from "react";

interface Props {
  textToCopy: string;
  className?: string;
}

export default function CopyButton({ textToCopy, className = "" }: Props) {
  const [isCopied, setIsCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(textToCopy);
      setIsCopied(true);

      setTimeout(() => {
        setIsCopied(false);
      }, 2000);
    } catch (error) {
      console.error("Error copying to clipboard:", error);
    }
  };

  return (
    <button
      onClick={handleCopy}
      className={`${className} ${isCopied ? "copied" : ""}`}
      disabled={isCopied}
    >
      {isCopied ? "Copied!" : "Copy"}
    </button>
  );
}
