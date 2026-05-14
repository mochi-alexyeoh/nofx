import { useState, useEffect, useRef } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Menu, X, ChevronDown, Settings, Linkedin, Facebook, Instagram } from 'lucide-react'
import { t, type Language } from '../../i18n/translations'
import {
  getPostAuthPath,
  getUserMode,
  setUserMode,
  type UserMode,
} from '../../lib/onboarding'
import { getCurrentPageForPath, ROUTES, type Page } from '../../router/paths'

interface HeaderBarProps {
  onLoginClick?: () => void
  isLoggedIn?: boolean
  isHomePage?: boolean
  currentPage?: Page
  language?: Language
  onLanguageChange?: (lang: Language) => void
  user?: { email: string } | null
  onLogout?: () => void
  onPageChange?: (page: Page) => void
  onLoginRequired?: (featureName: string) => void
}

export default function HeaderBar({
  isLoggedIn = false,
  isHomePage: _isHomePage = false,
  currentPage,
  language = 'zh' as Language,
  onLanguageChange,
  user,
  onLogout,
  onPageChange,
  onLoginRequired,
}: HeaderBarProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [languageDropdownOpen, setLanguageDropdownOpen] = useState(false)
  const [userDropdownOpen, setUserDropdownOpen] = useState(false)
  const [userMode, setUserModeState] = useState<UserMode>(
    () => getUserMode() ?? 'advanced'
  )
  const dropdownRef = useRef<HTMLDivElement>(null)
  const userDropdownRef = useRef<HTMLDivElement>(null)
  const resolvedCurrentPage =
    currentPage ?? getCurrentPageForPath(location.pathname)

  const navigateInApp = (path: string) => {
    navigate(path)
  }

  const handleSwitchMode = (nextMode: UserMode) => {
    setUserMode(nextMode)
    setUserModeState(nextMode)
    setUserDropdownOpen(false)
    navigateInApp(getPostAuthPath(nextMode))
  }
  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setLanguageDropdownOpen(false)
      }
      if (
        userDropdownRef.current &&
        !userDropdownRef.current.contains(event.target as Node)
      ) {
        setUserDropdownOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])

  useEffect(() => {
    if (mobileMenuOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }

    return () => {
      document.body.style.overflow = ''
    }
  }, [mobileMenuOpen])

  return (
    <nav className="fixed top-0 w-full z-50 header-bar">
      <div className="flex items-center justify-between h-16 px-4 sm:px-6 max-w-[1920px] mx-auto">
        {/* Logo - Always go to home page */}
        <div
          onClick={() => {
            navigateInApp(ROUTES.home)
          }}
          className="flex items-center hover:opacity-80 transition-opacity cursor-pointer"
        >
          <img
            src="/icons/zessai.png"
            alt="Zess Logo"
            className="h-10 w-[180px] object-contain object-left"
          />
        </div>

        {/* Desktop Menu */}
        <div className="hidden md:flex items-center justify-between flex-1 ml-8">
          {/* Left Side - Navigation Tabs - Always show all tabs */}
          <div className="flex items-center gap-2">
            {/* Navigation tabs configuration */}
            {(() => {
              // Define all navigation tabs
              const navTabs: {
                page: Page
                path: string
                label: string
                requiresAuth: boolean
                badge?: string
                hidden?: boolean
              }[] = [
                  {
                    page: 'agent',
                    path: ROUTES.agent,
                    label: 'Agent',
                    requiresAuth: false,
                    badge: 'Beta',
                    hidden: true,
                  },
                  {
                                      },
                  {
                    page: 'strategy-market',
                    path: ROUTES.strategyMarket,
                    label:
                      language === 'zh'
                        ? '策略市场'
                        : language === 'id'
                          ? 'Pasar'
                          : 'Market',
                    requiresAuth: true,
                  },
                  {
                    page: 'traders',
                    path: ROUTES.traders,
                    label: t('configNav', language),
                    requiresAuth: true,
                  },
                  {
                    page: 'trader',
                    path: ROUTES.dashboard,
                    label: t('dashboardNav', language),
                    requiresAuth: true,
                  },
                  {
                    page: 'strategy',
                    path: ROUTES.strategy,
                    label: t('strategyNav', language),
                    requiresAuth: true,
                  },
                  {
                    page: 'backtest',
                    path: ROUTES.backtest,
                    label: language === 'zh' ? '回测' : language === 'id' ? 'Backtest' : 'Backtest',
                    requiresAuth: true,
                  },
                  {
                    page: 'competition',
                    path: ROUTES.competition,
                    label: t('realtimeNav', language),
                    requiresAuth: true,
                  },
                  {
                    page: 'faq',
                    path: ROUTES.faq,
                    label: t('faqNav', language),
                    requiresAuth: false,
                  },
                ]

              const handleNavClick = (tab: (typeof navTabs)[0]) => {
                // If requires auth and not logged in, show login prompt
                if (tab.requiresAuth && !isLoggedIn) {
                  onLoginRequired?.(tab.label)
                  return
                }
                // Navigate normally
                if (onPageChange) {
                  onPageChange(tab.page)
                }
                navigateInApp(tab.path)
              }

              return navTabs.filter((tab) => !tab.hidden).map((tab) => (
                <button
                  key={tab.page}
                  onClick={() => handleNavClick(tab)}
                  className={`text-sm font-bold transition-all duration-300 relative focus:outline-2 focus:outline-yellow-500 px-3 py-2 rounded-lg
                    ${resolvedCurrentPage === tab.page ? 'text-nofx-gold' : 'text-nofx-text-muted hover:text-nofx-gold'}`}
                >
                  {resolvedCurrentPage === tab.page && (
                    <span className="absolute inset-0 rounded-lg bg-nofx-gold/15 -z-10" />
                  )}
                  {tab.label}
                  {tab.badge && (
                    <span className="ml-1 text-[10px] px-1.5 py-0.5 rounded-full bg-nofx-gold/20 text-nofx-gold font-semibold uppercase align-top relative -top-1">
                      {tab.badge}
                    </span>
                  )}
                </button>
              ))
            })()}
          </div>

          {/* Right Side - Social Links and User Actions */}
          <div className="flex items-center gap-4">
            {/* Social Links - Always visible */}
            <div className="flex items-center gap-1">
              <a
                href="https://www.linkedin.com/company/ntglobal-nextech/"
                target="_blank"
                rel="noopener noreferrer"
                className="p-2 rounded-lg transition-all hover:scale-110 text-nofx-text-muted hover:text-[#0A66C2] hover:bg-[#0A66C2]/10"
                title="LinkedIn"
              >
                <Linkedin size={16} />
              </a>
              <a
                href="https://www.facebook.com/NexTechPage"
                target="_blank"
                rel="noopener noreferrer"
                className="p-2 rounded-lg transition-all hover:scale-110 text-nofx-text-muted hover:text-[#1877F2] hover:bg-[#1877F2]/10"
                title="Facebook"
              >
                <Facebook size={16} />
              </a>
              <a
                href="https://www.instagram.com/nextechofficial_"
                target="_blank"
                rel="noopener noreferrer"
                className="p-2 rounded-lg transition-all hover:scale-110 text-nofx-text-muted hover:text-[#E4405F] hover:bg-[#E4405F]/10"
                title="Instagram"
              >
                <Instagram size={16} />
              </a>
            </div>

            {/* Divider */}
            <div className="h-5 w-px" style={{ background: '#2B3139' }} />

            {/* User Info and Actions */}
            {isLoggedIn && user ? (
              <div className="flex items-center gap-3">
                {/* User Info with Dropdown */}
                <div className="relative" ref={userDropdownRef}>
                  <button
                    onClick={() => setUserDropdownOpen(!userDropdownOpen)}
                    className="flex items-center gap-2 px-3 py-2 rounded transition-colors bg-nofx-bg-lighter border border-nofx-gold/20 hover:bg-white/5"
                  >
                    <div className="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold bg-nofx-gold text-black">
                      {user.email[0].toUpperCase()}
                    </div>
                    <span className="text-sm text-nofx-text-muted">
                      {user.email}
                    </span>
                    <ChevronDown className="w-4 h-4 text-nofx-text-muted" />
                  </button>

                  {userDropdownOpen && (
                    <div className="absolute right-0 top-full mt-2 w-48 rounded-lg shadow-lg overflow-hidden z-50 bg-nofx-bg-lighter border border-nofx-gold/20">
                      <div className="px-3 py-2 border-b border-nofx-gold/20">
                        <div className="text-xs text-nofx-text-muted">
                          {t('loggedInAs', language)}
                        </div>
                        <div className="text-sm font-medium text-nofx-text-muted">
                          {user.email}
                        </div>
                      </div>
                      <button
                        onClick={() => {
                          navigateInApp(ROUTES.settings)
                          setUserDropdownOpen(false)
                        }}
                        className="w-full flex items-center gap-2 px-3 py-2 text-sm transition-colors hover:bg-white/5 text-nofx-text-muted hover:text-white"
                      >
                        <Settings className="w-3.5 h-3.5" />
                        Settings
                      </button>
                      <button
                        onClick={() =>
                          handleSwitchMode(
                            userMode === 'beginner' ? 'advanced' : 'beginner'
                          )
                        }
                        className="w-full flex items-center gap-2 px-3 py-2 text-sm transition-colors hover:bg-white/5 text-nofx-text-muted hover:text-white"
                      >
                        <Settings className="w-3.5 h-3.5" />
                        {userMode === 'beginner'
                          ? language === 'zh'
                            ? '切到老手模式'
                            : 'Switch to Advanced'
                          : language === 'zh'
                            ? '切到新手模式'
                            : 'Switch to Beginner'}
                      </button>
                      {onLogout && (
                        <button
                          onClick={() => {
                            onLogout()
                            setUserDropdownOpen(false)
                          }}
                          className="w-full px-3 py-2 text-sm font-semibold transition-colors hover:opacity-80 text-center bg-nofx-danger/20 text-nofx-danger"
                        >
                          {t('exitLogin', language)}
                        </button>
                      )}
                    </div>
                  )}
                </div>
              </div>
            ) : (
              /* Show login/register buttons when not logged in and not on login/register pages */
              resolvedCurrentPage !== 'login' &&
              resolvedCurrentPage !== 'register' && (
                <div className="flex items-center gap-3">
                  <button
                    type="button"
                    onClick={() => navigateInApp(ROUTES.login)}
                    className="px-3 py-2 text-sm font-medium transition-colors rounded text-nofx-text-muted hover:text-white"
                  >
                    {t('signIn', language)}
                  </button>
                  <button
                    type="button"
                    onClick={() => navigateInApp(ROUTES.register)}
                    className="px-3 py-2 text-sm font-semibold transition-colors rounded bg-nofx-gold text-black hover:bg-yellow-400"
                  >
                    {t('signUp', language) || 'Sign Up'}
                  </button>
                </div>
              )
            )}

            {/* Language Toggle - Always at the rightmost */}
            <div className="relative" ref={dropdownRef}>
              <button
                onClick={() => setLanguageDropdownOpen(!languageDropdownOpen)}
                className="flex items-center gap-2 px-3 py-2 rounded transition-colors text-nofx-text-muted hover:bg-white/5"
              >
                <span className="text-lg">
                  {language === 'zh' ? '🇨🇳' : language === 'id' ? '🇮🇩' : '🇺🇸'}
                </span>
                <ChevronDown className="w-4 h-4" />
              </button>

              {languageDropdownOpen && (
                <div className="absolute right-0 top-full mt-2 w-32 rounded-lg shadow-lg overflow-hidden z-50 bg-nofx-bg-lighter border border-nofx-gold/20">
                  <button
                    onClick={() => {
                      onLanguageChange?.('zh')
                      setLanguageDropdownOpen(false)
                    }}
                    className={`w-full flex items-center gap-2 px-3 py-2 transition-colors text-nofx-text-muted hover:text-white
                      ${language === 'zh' ? 'bg-nofx-gold/10' : 'hover:bg-white/5'}`}
                  >
                    <span className="text-base">🇨🇳</span>
                    <span className="text-sm">中文</span>
                  </button>
                  <button
                    onClick={() => {
                      onLanguageChange?.('en')
                      setLanguageDropdownOpen(false)
                    }}
                    className={`w-full flex items-center gap-2 px-3 py-2 transition-colors text-nofx-text-muted hover:text-white
                      ${language === 'en' ? 'bg-nofx-gold/10' : 'hover:bg-white/5'}`}
                  >
                    <span className="text-base">🇺🇸</span>
                    <span className="text-sm">English</span>
                  </button>
                  <button
                    onClick={() => {
                      onLanguageChange?.('id')
                      setLanguageDropdownOpen(false)
                    }}
                    className={`w-full flex items-center gap-2 px-3 py-2 transition-colors text-nofx-text-muted hover:text-white
                      ${language === 'id' ? 'bg-nofx-gold/10' : 'hover:bg-white/5'}`}
                  >
                    <span className="text-base">🇮🇩</span>
                    <span className="text-sm">Bahasa</span>
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Mobile Menu Button */}
        <motion.button
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          className="md:hidden text-nofx-text-muted hover:text-white"
          whileTap={{ scale: 0.9 }}
        >
          {mobileMenuOpen ? (
            <X className="w-6 h-6" />
          ) : (
            <Menu className="w-6 h-6" />
          )}
        </motion.button>
      </div>

      {/* Mobile Menu Overlay (always mounted for instant open on heavy landing page) */}
      <div
        className={`fixed inset-0 z-40 md:hidden transition-opacity duration-100 ${mobileMenuOpen ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none'}`}
        style={{ top: '64px' }} // Below header
      >
        <button
          type="button"
          aria-label="Close menu"
          onClick={() => setMobileMenuOpen(false)}
          className="absolute inset-0 bg-black/75"
        />
        <div
          className={`relative z-10 flex flex-col h-[calc(100vh-64px)] w-[86vw] max-w-sm overflow-y-auto px-5 py-6 bg-[#0B0E11] border-r border-zinc-800 shadow-2xl transition-transform duration-100 ${mobileMenuOpen ? 'translate-x-0' : '-translate-x-full'}`}
        >
          {/* Navigation Links */}
          <div className="flex flex-col gap-4 mb-5">
            {(() => {
              const navTabs: {
                page: Page
                path: string
                label: string
                requiresAuth: boolean
                badge?: string
                hidden?: boolean
              }[] = [
                  {
                    page: 'agent',
                    path: ROUTES.agent,
                    label: 'Agent',
                    requiresAuth: false,
                    badge: 'Beta',
                    hidden: true,
                  },
                  {
                                      },
                  {
                    page: 'strategy-market',
                    path: ROUTES.strategyMarket,
                    label:
                      language === 'zh'
                        ? '策略市场'
                        : language === 'id'
                          ? 'Pasar'
                          : 'Market',
                    requiresAuth: true,
                  },
                  {
                    page: 'traders',
                    path: ROUTES.traders,
                    label: t('configNav', language),
                    requiresAuth: true,
                  },
                  {
                    page: 'trader',
                    path: ROUTES.dashboard,
                    label: t('dashboardNav', language),
                    requiresAuth: true,
                  },
                  {
                    page: 'strategy',
                    path: ROUTES.strategy,
                    label: t('strategyNav', language),
                    requiresAuth: true,
                  },
                  {
                    page: 'backtest',
                    path: ROUTES.backtest,
                    label: language === 'zh' ? '回测' : language === 'id' ? 'Backtest' : 'Backtest',
                    requiresAuth: true,
                  },
                  {
                    page: 'competition',
                    path: ROUTES.competition,
                    label: t('realtimeNav', language),
                    requiresAuth: true,
                  },
                  {
                    page: 'faq',
                    path: ROUTES.faq,
                    label: t('faqNav', language),
                    requiresAuth: false,
                  },
                ]

              const handleMobileNavClick = (tab: (typeof navTabs)[0]) => {
                if (tab.requiresAuth && !isLoggedIn) {
                  onLoginRequired?.(tab.label)
                  setMobileMenuOpen(false)
                  return
                }
                if (onPageChange) {
                  onPageChange(tab.page)
                }
                navigateInApp(tab.path)
                setMobileMenuOpen(false)
              }

              return navTabs.filter((tab) => !tab.hidden).map((tab) => (
                <button
                  key={tab.page}
                  onClick={() => handleMobileNavClick(tab)}
                  className={`w-full rounded-lg px-3 py-2 text-xl font-extrabold tracking-tight text-left flex items-center gap-3 transition-colors
                        ${resolvedCurrentPage === tab.page ? 'text-nofx-gold bg-nofx-gold/10 border border-nofx-gold/20' : 'text-zinc-200 hover:bg-white/5'}`}
                >
                  {resolvedCurrentPage === tab.page && (
                    <div className="w-1.5 h-1.5 rounded-full bg-nofx-gold" />
                  )}
                  {tab.label}
                  {tab.badge && (
                    <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-nofx-gold/20 text-nofx-gold font-semibold uppercase align-middle relative -top-1">
                      {tab.badge}
                    </span>
                  )}
                  {tab.requiresAuth && !isLoggedIn && (
                    <span className="text-[10px] px-1.5 py-0.5 rounded border border-zinc-800 text-zinc-500 font-normal tracking-wide uppercase align-middle relative -top-1">
                      LOGIN_REQ
                    </span>
                  )}
                </button>
              ))
            })()}

          </div>

          {/* Bottom Actions */}
          <div className="mt-6 space-y-5">
            {/* Social Links */}
            <div className="flex items-center gap-4">
              {[
                {
                  href: 'https://www.linkedin.com/company/ntglobal-nextech/',
                  icon: Linkedin,
                  hover: 'hover:text-[#0A66C2] hover:border-[#0A66C2]'
                },
                {
                  href: 'https://www.facebook.com/NexTechPage',
                  icon: Facebook,
                  hover: 'hover:text-[#1877F2] hover:border-[#1877F2]'
                },
                {
                  href: 'https://www.instagram.com/nextechofficial_',
                  icon: Instagram,
                  hover: 'hover:text-[#E4405F] hover:border-[#E4405F]'
                },
              ].map((link, i) => {
                const Icon = link.icon
                return (
                  <a
                    key={i}
                    href={link.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className={`w-12 h-12 rounded-full bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-500 transition-colors ${link.hover}`}
                  >
                    <Icon size={20} />
                  </a>
                )
              })}
            </div>

            {/* Account / Lang */}
            <div className="grid grid-cols-2 gap-4">
              {/* Lang Switcher */}
              <div className="flex bg-zinc-900 rounded-lg p-1 border border-zinc-800">
                {['zh', 'en', 'id'].map((lang) => (
                  <button
                    key={lang}
                    onClick={() => {
                      onLanguageChange?.(lang as Language)
                      setMobileMenuOpen(false)
                    }}
                    className={`flex-1 py-3 text-sm font-bold rounded-md transition-colors ${language === lang
                      ? 'bg-zinc-800 text-white shadow-sm'
                      : 'text-zinc-500'
                      }`}
                  >
                    {lang === 'zh' ? 'CN' : lang === 'id' ? 'ID' : 'EN'}
                  </button>
                ))}
              </div>

              {/* Auth Actions */}
              {isLoggedIn && user ? (
                <button
                  onClick={() => {
                    onLogout?.()
                    setMobileMenuOpen(false)
                  }}
                  className="bg-red-500/10 border border-red-500/20 text-red-500 rounded-lg font-bold text-sm hover:bg-red-500/20 transition-colors"
                >
                  {t('exitLogin', language)}
                </button>
              ) : (
                resolvedCurrentPage !== 'login' &&
                resolvedCurrentPage !== 'register' && (
                  <div className="grid grid-cols-2 gap-2">
                    <button
                      type="button"
                      onClick={() => {
                        navigateInApp(ROUTES.login)
                        setMobileMenuOpen(false)
                      }}
                      className="flex items-center justify-center bg-nofx-gold text-black rounded-lg font-bold text-sm hover:bg-yellow-400 transition-colors"
                    >
                      {t('signIn', language)}
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        navigateInApp(ROUTES.register)
                        setMobileMenuOpen(false)
                      }}
                      className="flex items-center justify-center border border-zinc-700 text-zinc-200 rounded-lg font-bold text-sm hover:bg-zinc-800 transition-colors"
                    >
                      {t('signUp', language) || 'Sign Up'}
                    </button>
                  </div>
                )
              )}
            </div>
          </div>
        </div>
      </div>
    </nav>
  )
}
