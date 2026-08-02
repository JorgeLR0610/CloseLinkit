import { 
  FaFacebook, 
  FaLinkedin, 
  FaTelegram, 
  FaWhatsapp, 
  FaXTwitter 
} from 'react-icons/fa6'
import CopyButton from '../CopyButton/CopyButton'
import './RecentURLBox.css'
import ShareButton from '../ShareButton/ShareButton'

export default function RecentURLBox({ shortURL }: { shortURL: string }) {
  const encodedURL = encodeURIComponent(shortURL)
  const whatsappURL =
      `https://api.whatsapp.com/send?text=${encodedURL}`

  const facebookURL =
      `https://www.facebook.com/sharer/sharer.php?u=${encodedURL}`

  const xTwitterURL =
      `https://x.com/intent/post?text=${encodeURIComponent(
        "Check out this link!"
      )}&url=${encodedURL}`

  const telegramURL =
      `https://t.me/share/url?url=${encodedURL}`

  const linkedinURL =
    `https://www.linkedin.com/sharing/share-offsite/?url=${encodedURL}`

  return (
    <div className="recent-box glass-panel fade-in">
      <div className="recent-top">
        <span className="recent-url">{shortURL}</span>
        <CopyButton textToCopy={shortURL} className='copy-btn'/>
      </div>
      <div className="recent-bottom">
        <div className="social-icons">
          <ShareButton
            href={whatsappURL}
            icon={FaWhatsapp}
            label="WhatsApp"
            className='whatsapp'
          />
          <ShareButton
            href={facebookURL}
            icon={FaFacebook}
            label="facebook"
            className='facebook'
          />
          <ShareButton
            href={linkedinURL}
            icon={FaLinkedin}
            label="LinkedIn"
            className='linkedin'
          />
          <ShareButton
            href={telegramURL}
            icon={FaTelegram}
            label="Telegram"
            className='telegram'
          />
          <ShareButton
            href={xTwitterURL}
            icon={FaXTwitter}
            label="X"
            className='x'
          />
        </div>
      </div>
    </div>
  )
}
