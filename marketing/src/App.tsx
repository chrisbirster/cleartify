import { useState, type ReactNode, type SVGProps } from 'react'

type IconName =
  | 'arrow-right'
  | 'bot'
  | 'brain'
  | 'check'
  | 'chevron-down'
  | 'code'
  | 'download'
  | 'eye'
  | 'file-code'
  | 'github'
  | 'inbox'
  | 'layers'
  | 'linkedin'
  | 'menu'
  | 'minus'
  | 'palette'
  | 'rocket'
  | 'server'
  | 'twitter'
  | 'users'
  | 'x'

type NavLink = {
  label: string
  href: string
}

type FeatureCard = {
  icon: IconName
  title: string
  description: string
}

type Step = {
  number: number
  icon: IconName
  title: string
  description: string
}

type PreviewTone = 'accent' | 'info' | 'muted' | 'warning'

type DemoTab = {
  id: string
  label: string
  icon: IconName
  title: string
  badge: string
  items: Array<{
    title: string
    meta: string
    tone: PreviewTone
  }>
}

type MatrixValue = 'check' | 'minus' | string

type MatrixRow = {
  feature: string
  traditional: MatrixValue
  genericAI: MatrixValue
  cleartify: MatrixValue
}

type AudienceCard = {
  icon: IconName
  title: string
  description: string
}

type Testimonial = {
  initials: string
  quote: string
  name: string
  role: string
  company: string
}

type PricingTier = {
  name: string
  price: string
  period?: string
  description: string
  features: string[]
  ctaLabel: string
  ctaHref: string
  featured?: boolean
}

type FAQItem = {
  question: string
  answer: string
}

type FooterColumn = {
  title: string
  links: Array<{
    label: string
    href: string
  }>
}

const appOrigin = (
  import.meta.env.VITE_APP_ORIGIN ??
  (import.meta.env.DEV ? getDevAppOrigin() : 'https://app.cleartify.com')
).replace(/\/$/, '')

const signupHref = `${appOrigin}/signup`
const loginHref = `${appOrigin}/login`
const demoHref = '#demo'

function getDevAppOrigin() {
  return 'http://localhost:3000'
}

const navLinks: NavLink[] = [
  { label: 'Product', href: '#product' },
  { label: 'How it Works', href: '#how-it-works' },
  { label: 'Agents', href: '#agents' },
  { label: 'Pricing', href: '#pricing' },
  { label: 'FAQ', href: '#faq' },
]

const heroBadges = [
  { icon: 'inbox' as const, label: 'Context Inbox' },
  { icon: 'layers' as const, label: 'Prototype-to-Plan' },
  { icon: 'bot' as const, label: 'Agent Execution' },
  { icon: 'brain' as const, label: 'Product Memory' },
]

const logos = ['Acme Corp', 'Vertex AI', 'Quantum Labs', 'NovaTech', 'Stellar Systems', 'Cipher Works']

const oldWorldItems = [
  'Specs become tickets',
  'Tickets lose context',
  'Teams manage process',
  'Humans do repetitive coordination',
]

const newWorldItems = [
  'Context stays attached to work',
  'Prototypes guide execution',
  'Agents handle procedural tasks',
  'Humans focus on judgment and taste',
]

const featureCards: FeatureCard[] = [
  {
    icon: 'inbox',
    title: 'Context Inbox',
    description: 'Collect feedback, notes, Slack threads, issues, and docs in one place.',
  },
  {
    icon: 'layers',
    title: 'Work Objects',
    description: 'Combine goals, evidence, prototypes, code, and decisions into unified work units.',
  },
  {
    icon: 'file-code',
    title: 'Prototype-to-Plan',
    description: 'Turn rough mocks and designs into detailed implementation plans automatically.',
  },
  {
    icon: 'bot',
    title: 'Agent Execution',
    description: 'Generate tasks, docs, tests, and code drafts with AI agents grounded in context.',
  },
  {
    icon: 'brain',
    title: 'Product Memory',
    description: 'Keep the "why" connected to what shipped. Never lose context again.',
  },
  {
    icon: 'users',
    title: 'Review Loop',
    description: 'Humans review judgment-heavy decisions before release. Stay in control.',
  },
]

const steps: Step[] = [
  {
    number: 1,
    icon: 'download',
    title: 'Bring in signals',
    description: 'Import customer feedback, docs, notes, GitHub issues, and prototypes from all your sources.',
  },
  {
    number: 2,
    icon: 'layers',
    title: 'Build shared context',
    description: 'Cleartify clusters the signal into a work object with the full story attached.',
  },
  {
    number: 3,
    icon: 'bot',
    title: 'Let agents execute',
    description: 'Agents draft plans, tasks, tests, docs, and code changes based on real context.',
  },
  {
    number: 4,
    icon: 'rocket',
    title: 'Review and ship',
    description: 'Humans guide, approve, and push work to production with full visibility.',
  },
]

const demoTabs: DemoTab[] = [
  {
    id: 'inbox',
    label: 'Inbox',
    icon: 'inbox',
    title: 'Context Inbox',
    badge: '4 new signals',
    items: [
      { title: 'Mobile performance issues reported', meta: 'feedback - 2m ago', tone: 'accent' },
      { title: 'Design review thread from #product', meta: 'slack - 15m ago', tone: 'accent' },
      { title: 'PR #234 needs context from original spec', meta: 'github - 1h ago', tone: 'muted' },
      { title: 'Competitor analysis notes from team sync', meta: 'note - 3h ago', tone: 'muted' },
    ],
  },
  {
    id: 'work-object',
    label: 'Work Object',
    icon: 'layers',
    title: 'Mobile Optimization Work Object',
    badge: 'Context attached',
    items: [
      { title: 'Goal: improve dashboard mobile render time', meta: 'objective', tone: 'accent' },
      { title: 'Evidence: 7 user reports and 2 benchmark traces', meta: 'research', tone: 'info' },
      { title: 'Prototype: revised chart loading states', meta: 'design', tone: 'accent' },
      { title: 'Decision: ship virtualization before chart rewrite', meta: 'decision', tone: 'warning' },
    ],
  },
  {
    id: 'prototype-review',
    label: 'Prototype Review',
    icon: 'eye',
    title: 'Prototype Review',
    badge: '5 screens attached',
    items: [
      { title: 'Onboarding screen 1 approved with copy edits', meta: 'review', tone: 'accent' },
      { title: 'Progress indicator updated for mobile', meta: 'interaction', tone: 'info' },
      { title: 'Consent step needs accessibility review', meta: 'design QA', tone: 'warning' },
      { title: 'Desktop variant linked to implementation plan', meta: 'handoff', tone: 'muted' },
    ],
  },
  {
    id: 'agent-workspace',
    label: 'Agent Workspace',
    icon: 'bot',
    title: 'Agent Workspace',
    badge: '3 agents running',
    items: [
      { title: 'Drafting implementation plan from context bundle', meta: 'planning agent', tone: 'accent' },
      { title: 'Generating regression test checklist', meta: 'quality agent', tone: 'info' },
      { title: 'Preparing code review summary for PR #234', meta: 'review agent', tone: 'accent' },
      { title: 'Waiting on human approval for rollout step', meta: 'approval gate', tone: 'warning' },
    ],
  },
  {
    id: 'code-review',
    label: 'Code Review',
    icon: 'code',
    title: 'Code Review',
    badge: '2 findings',
    items: [
      { title: 'Chart memoization regression in mobile breakpoint path', meta: 'performance', tone: 'warning' },
      { title: 'Spec link now embedded directly into review context', meta: 'traceability', tone: 'accent' },
      { title: 'Loading skeleton updated to match approved prototype', meta: 'ui parity', tone: 'info' },
      { title: 'Release note drafted from work object timeline', meta: 'handoff', tone: 'muted' },
    ],
  },
]

const matrixRows: MatrixRow[] = [
  {
    feature: 'Source of truth',
    traditional: 'Tickets',
    genericAI: 'Chat history',
    cleartify: 'Shared context layer',
  },
  {
    feature: 'Prototype support',
    traditional: 'minus',
    genericAI: 'minus',
    cleartify: 'check',
  },
  {
    feature: 'AI agents',
    traditional: 'minus',
    genericAI: 'check',
    cleartify: 'check',
  },
  {
    feature: 'Shared context',
    traditional: 'minus',
    genericAI: 'minus',
    cleartify: 'check',
  },
  {
    feature: 'Code-aware workflow',
    traditional: 'minus',
    genericAI: 'check',
    cleartify: 'check',
  },
  {
    feature: 'Product memory',
    traditional: 'minus',
    genericAI: 'minus',
    cleartify: 'check',
  },
]

const audiences: AudienceCard[] = [
  {
    icon: 'rocket',
    title: 'Founder-led startups',
    description: 'Move fast without losing the thread. Keep product decisions connected to shipped code.',
  },
  {
    icon: 'code',
    title: 'Product engineers',
    description: 'Spend less time on ticket grooming. More time building what matters.',
  },
  {
    icon: 'palette',
    title: 'Design engineers',
    description: 'Go from prototype to production with context intact. No more lost design decisions.',
  },
  {
    icon: 'server',
    title: 'Internal platform teams',
    description: 'Build internal tools faster with agents that understand your system context.',
  },
  {
    icon: 'users',
    title: 'Small product teams',
    description: 'Replace ticket-heavy workflows with a system built for speed and clarity.',
  },
]

const testimonials: Testimonial[] = [
  {
    initials: 'SC',
    quote:
      'Cleartify changed how we think about tickets. Now context flows from idea to shipped code without the usual friction.',
    name: 'Sarah Chen',
    role: 'Head of Product',
    company: 'NovaTech',
  },
  {
    initials: 'MR',
    quote:
      "The agents actually understand what we're building. They're grounded in our prototypes and decisions, not just code.",
    name: 'Marcus Rodriguez',
    role: 'Product Engineer',
    company: 'Vertex AI',
  },
  {
    initials: 'AK',
    quote: 'We shipped our last three features in half the time. Less ticket theater, more building.',
    name: 'Alex Kim',
    role: 'Founder',
    company: 'Quantum Labs',
  },
]

const pricingTiers: PricingTier[] = [
  {
    name: 'Starter',
    price: '$0',
    period: '/month',
    description: 'For individuals exploring context-driven development.',
    features: ['Up to 3 work objects', 'Basic context inbox', '1 AI agent session/day', 'Community support'],
    ctaLabel: 'Get started free',
    ctaHref: signupHref,
  },
  {
    name: 'Team',
    price: '$29',
    period: '/user/month',
    description: 'For teams shipping fast with shared context.',
    features: [
      'Unlimited work objects',
      'Full context inbox',
      'Unlimited AI agent sessions',
      'Prototype-to-plan',
      'Product memory',
      'Priority support',
    ],
    ctaLabel: 'Start with a prototype',
    ctaHref: signupHref,
    featured: true,
  },
  {
    name: 'Enterprise',
    price: 'Custom',
    description: 'For organizations with advanced needs.',
    features: [
      'Everything in Team',
      'Custom integrations',
      'SSO & advanced security',
      'Dedicated success manager',
      'SLA & uptime guarantees',
      'On-premise option',
    ],
    ctaLabel: 'Contact sales',
    ctaHref: demoHref,
  },
]

const faqItems: FAQItem[] = [
  {
    question: 'Is Cleartify a replacement for Jira or Linear?',
    answer:
      'Cleartify is built for the layer before execution gets flattened into tickets. Teams use it to keep context, prototypes, evidence, and decisions attached to the work so agents and humans can execute with less translation loss.',
  },
  {
    question: 'How do agents use context?',
    answer:
      "Our agents are grounded in your product context: feedback, prototypes, decisions, and code. That means they understand the why behind every task and produce outputs that match your team's intent.",
  },
  {
    question: 'Can I start from a prototype instead of a spec?',
    answer:
      'Yes. Cleartify is designed for prototype-first teams. Rough mocks, screenshots, and design explorations can all feed the context layer and become structured implementation plans.',
  },
  {
    question: 'Does Cleartify integrate with GitHub and Slack?',
    answer:
      'Yes. The platform is designed to ingest product signals from the tools teams already use, including chat threads, issues, docs, and code review workflows.',
  },
  {
    question: 'Do humans stay in control of execution?',
    answer:
      'Always. Agents handle procedural work, but humans stay in the loop for judgment-heavy decisions, approvals, and release control.',
  },
]

const footerColumns: FooterColumn[] = [
  {
    title: 'Product',
    links: [
      { label: 'Features', href: '#product' },
      { label: 'Pricing', href: '#pricing' },
      { label: 'Integrations', href: '#demo' },
      { label: 'Changelog', href: '#demo' },
      { label: 'Roadmap', href: '#faq' },
    ],
  },
  {
    title: 'Company',
    links: [
      { label: 'About', href: '#agents' },
      { label: 'Blog', href: '#faq' },
      { label: 'Careers', href: '#agents' },
      { label: 'Press', href: '#demo' },
    ],
  },
  {
    title: 'Resources',
    links: [
      { label: 'Documentation', href: '#how-it-works' },
      { label: 'API Reference', href: '#demo' },
      { label: 'Community', href: '#faq' },
      { label: 'Support', href: loginHref },
    ],
  },
  {
    title: 'Legal',
    links: [
      { label: 'Privacy', href: '#faq' },
      { label: 'Terms', href: '#faq' },
      { label: 'Security', href: '#faq' },
    ],
  },
]

function cn(...parts: Array<string | false | null | undefined>) {
  return parts.filter(Boolean).join(' ')
}

function Icon({ name, className, ...props }: { name: IconName; className?: string } & SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      className={className}
      {...props}
    >
      {name === 'menu' ? (
        <>
          <path d="M4 5h16" />
          <path d="M4 12h16" />
          <path d="M4 19h16" />
        </>
      ) : null}
      {name === 'x' ? (
        <>
          <path d="M18 6 6 18" />
          <path d="m6 6 12 12" />
        </>
      ) : null}
      {name === 'inbox' ? (
        <>
          <polyline points="22 12 16 12 14 15 10 15 8 12 2 12" />
          <path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z" />
        </>
      ) : null}
      {name === 'layers' ? (
        <>
          <path d="M12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83z" />
          <path d="M2 12a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 12" />
          <path d="M2 17a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 17" />
        </>
      ) : null}
      {name === 'bot' ? (
        <>
          <path d="M12 8V4H8" />
          <rect width="16" height="12" x="4" y="8" rx="2" />
          <path d="M2 14h2" />
          <path d="M20 14h2" />
          <path d="M15 13v2" />
          <path d="M9 13v2" />
        </>
      ) : null}
      {name === 'brain' ? (
        <>
          <path d="M12 18V5" />
          <path d="M15 13a4.17 4.17 0 0 1-3-4 4.17 4.17 0 0 1-3 4" />
          <path d="M17.598 6.5A3 3 0 1 0 12 5a3 3 0 1 0-5.598 1.5" />
          <path d="M17.997 5.125a4 4 0 0 1 2.526 5.77" />
          <path d="M18 18a4 4 0 0 0 2-7.464" />
          <path d="M19.967 17.483A4 4 0 1 1 12 18a4 4 0 1 1-7.967-.517" />
          <path d="M6 18a4 4 0 0 1-2-7.464" />
          <path d="M6.003 5.125a4 4 0 0 0-2.526 5.77" />
        </>
      ) : null}
      {name === 'file-code' ? (
        <>
          <path d="M6 22a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h8a2.4 2.4 0 0 1 1.704.706l3.588 3.588A2.4 2.4 0 0 1 20 8v12a2 2 0 0 1-2 2z" />
          <path d="M14 2v5a1 1 0 0 0 1 1h5" />
          <path d="M10 12.5 8 15l2 2.5" />
          <path d="m14 12.5 2 2.5-2 2.5" />
        </>
      ) : null}
      {name === 'users' ? (
        <>
          <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
          <path d="M16 3.128a4 4 0 0 1 0 7.744" />
          <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
          <circle cx="9" cy="7" r="4" />
        </>
      ) : null}
      {name === 'download' ? (
        <>
          <path d="M12 15V3" />
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <path d="m7 10 5 5 5-5" />
        </>
      ) : null}
      {name === 'rocket' ? (
        <>
          <path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z" />
          <path d="m12 15-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z" />
          <path d="M9 12H4s.55-3.03 2-4c1.62-1.08 5 0 5 0" />
          <path d="M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5" />
        </>
      ) : null}
      {name === 'eye' ? (
        <>
          <path d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0" />
          <circle cx="12" cy="12" r="3" />
        </>
      ) : null}
      {name === 'code' ? (
        <>
          <path d="m16 18 6-6-6-6" />
          <path d="m8 6-6 6 6 6" />
        </>
      ) : null}
      {name === 'check' ? (
        <>
          <path d="M20 6 9 17l-5-5" />
        </>
      ) : null}
      {name === 'minus' ? (
        <>
          <path d="M5 12h14" />
        </>
      ) : null}
      {name === 'chevron-down' ? (
        <>
          <path d="m6 9 6 6 6-6" />
        </>
      ) : null}
      {name === 'palette' ? (
        <>
          <path d="M12 22a1 1 0 0 1 0-20 10 9 0 0 1 10 9 5 5 0 0 1-5 5h-2.25a1.75 1.75 0 0 0-1.4 2.8l.3.4a1.75 1.75 0 0 1-1.4 2.8z" />
          <circle cx="13.5" cy="6.5" r=".5" fill="currentColor" />
          <circle cx="17.5" cy="10.5" r=".5" fill="currentColor" />
          <circle cx="6.5" cy="12.5" r=".5" fill="currentColor" />
          <circle cx="8.5" cy="7.5" r=".5" fill="currentColor" />
        </>
      ) : null}
      {name === 'server' ? (
        <>
          <rect width="20" height="8" x="2" y="2" rx="2" ry="2" />
          <rect width="20" height="8" x="2" y="14" rx="2" ry="2" />
          <line x1="6" x2="6.01" y1="6" y2="6" />
          <line x1="6" x2="6.01" y1="18" y2="18" />
        </>
      ) : null}
      {name === 'arrow-right' ? (
        <>
          <path d="M5 12h14" />
          <path d="m12 5 7 7-7 7" />
        </>
      ) : null}
      {name === 'twitter' ? (
        <>
          <path d="M22 4s-.7 2.1-2 3.4c1.6 10-9.4 17.3-18 11.6 2.2.1 4.4-.6 6-2C3 15.5.5 9.6 3 5c2.2 2.6 5.6 4.1 9 4-.9-4.2 4-6.6 7-3.8 1.1 0 3-1.2 3-1.2z" />
        </>
      ) : null}
      {name === 'github' ? (
        <>
          <path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4" />
          <path d="M9 18c-4.51 2-5-2-7-2" />
        </>
      ) : null}
      {name === 'linkedin' ? (
        <>
          <path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-2-2 2 2 0 0 0-2 2v7h-4v-7a6 6 0 0 1 6-6z" />
          <rect width="4" height="12" x="2" y="9" />
          <circle cx="4" cy="4" r="2" />
        </>
      ) : null}
    </svg>
  )
}

function ButtonLink({
  href,
  children,
  variant = 'primary',
  className,
  onClick,
}: {
  href: string
  children: ReactNode
  variant?: 'ghost' | 'primary' | 'secondary'
  className?: string
  onClick?: () => void
}) {
  return (
    <a
      href={href}
      onClick={onClick}
      className={cn(
        'inline-flex items-center justify-center rounded-full text-sm font-medium transition-all duration-200',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
        variant === 'primary' && 'bg-primary px-6 py-3 text-primary-foreground hover:bg-primary/90',
        variant === 'secondary' &&
          'border border-border bg-background px-6 py-3 text-foreground shadow-sm hover:bg-accent hover:text-accent-foreground',
        variant === 'ghost' && 'px-2 py-2 text-muted-foreground hover:text-foreground',
        className,
      )}
    >
      {children}
    </a>
  )
}

function SectionHeading({
  title,
  description,
}: {
  title: string
  description: string
}) {
  return (
    <div className="mx-auto max-w-3xl text-center">
      <h2 className="text-balance text-3xl font-bold tracking-tight text-foreground sm:text-4xl">{title}</h2>
      <p className="mt-4 text-lg text-muted-foreground">{description}</p>
    </div>
  )
}

function MatrixCell({ value }: { value: MatrixValue }) {
  if (value === 'check') {
    return (
      <div className="flex justify-center">
        <div className="flex h-6 w-6 items-center justify-center rounded-full bg-accent/20">
          <Icon name="check" className="h-4 w-4 text-accent" />
        </div>
      </div>
    )
  }

  if (value === 'minus') {
    return (
      <div className="flex justify-center">
        <div className="flex h-6 w-6 items-center justify-center rounded-full bg-muted">
          <Icon name="minus" className="h-4 w-4 text-muted-foreground" />
        </div>
      </div>
    )
  }

  return (
    <div className="flex justify-center">
      <span className="text-sm text-muted-foreground">{value}</span>
    </div>
  )
}

function FAQEntry({
  item,
  open,
  onToggle,
}: {
  item: FAQItem
  open: boolean
  onToggle: () => void
}) {
  return (
    <div className="border-b border-border/50">
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={open}
        className="flex w-full items-start justify-between gap-4 py-4 text-left text-sm font-medium text-foreground transition-colors hover:text-foreground"
      >
        <span>{item.question}</span>
        <Icon
          name="chevron-down"
          className={cn('mt-0.5 h-4 w-4 shrink-0 text-muted-foreground transition-transform', open && 'rotate-180')}
        />
      </button>
      <div
        className={cn(
          'grid transition-[grid-template-rows,opacity] duration-200',
          open ? 'grid-rows-[1fr] opacity-100' : 'grid-rows-[0fr] opacity-0',
        )}
      >
        <div className="overflow-hidden">
          <div className="pb-4 text-sm leading-relaxed text-muted-foreground">{item.answer}</div>
        </div>
      </div>
    </div>
  )
}

export function App() {
  const [menuOpen, setMenuOpen] = useState(false)
  const [activeDemoId, setActiveDemoId] = useState(demoTabs[0].id)
  const [openQuestion, setOpenQuestion] = useState(faqItems[1].question)

  const activeDemoTab = demoTabs.find((tab) => tab.id === activeDemoId) ?? demoTabs[0]

  return (
    <div className="marketing-shell font-sans antialiased">
      <main className="min-h-screen">
        <header className="fixed top-0 left-0 right-0 z-50 border-b border-border/50 bg-background/80 backdrop-blur-xl">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="flex h-16 items-center justify-between">
              <a href="/" className="flex items-center gap-2" onClick={() => setMenuOpen(false)}>
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
                  <span className="text-sm font-bold text-primary-foreground">C</span>
                </div>
                <span className="text-lg font-semibold tracking-tight text-foreground">Cleartify</span>
              </a>

              <nav className="hidden items-center gap-8 md:flex">
                {navLinks.map((link) => (
                  <a
                    key={link.href}
                    href={link.href}
                    className="text-sm text-muted-foreground transition-colors hover:text-foreground"
                  >
                    {link.label}
                  </a>
                ))}
              </nav>

              <div className="hidden items-center gap-4 md:flex">
                <a href={demoHref} className="text-sm text-muted-foreground transition-colors hover:text-foreground">
                  See demo
                </a>
                <a href={loginHref} className="text-sm text-muted-foreground transition-colors hover:text-foreground">
                  Sign in
                </a>
                <ButtonLink href={signupHref} className="px-4 py-2">
                  Start with a prototype
                </ButtonLink>
              </div>

              <button
                type="button"
                className="md:hidden"
                aria-label="Toggle menu"
                aria-expanded={menuOpen}
                onClick={() => setMenuOpen((open) => !open)}
              >
                <Icon name={menuOpen ? 'x' : 'menu'} className="h-6 w-6 text-foreground" />
              </button>
            </div>
          </div>

          <div
            className={cn(
              'border-t border-border/50 bg-background/95 backdrop-blur-xl transition-[max-height,opacity] duration-200 md:hidden',
              menuOpen ? 'max-h-96 opacity-100' : 'max-h-0 overflow-hidden opacity-0',
            )}
          >
            <div className="mx-auto flex max-w-7xl flex-col gap-2 px-4 py-4 sm:px-6">
              {navLinks.map((link) => (
                <a
                  key={link.href}
                  href={link.href}
                  onClick={() => setMenuOpen(false)}
                  className="rounded-xl px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                >
                  {link.label}
                </a>
              ))}
              <div className="mt-2 flex flex-col gap-3">
                <ButtonLink href={demoHref} variant="secondary" onClick={() => setMenuOpen(false)}>
                  See demo
                </ButtonLink>
                <ButtonLink href={loginHref} variant="secondary" onClick={() => setMenuOpen(false)}>
                  Sign in
                </ButtonLink>
                <ButtonLink href={signupHref} onClick={() => setMenuOpen(false)}>
                  Start with a prototype
                </ButtonLink>
              </div>
            </div>
          </div>
        </header>

        <section className="relative overflow-hidden pt-32 pb-20 lg:pt-40 lg:pb-32">
          <div className="pointer-events-none absolute inset-0 overflow-hidden">
            <div className="absolute -top-1/2 left-1/2 h-[800px] w-[800px] -translate-x-1/2 rounded-full bg-accent/5 blur-3xl" />
            <div className="absolute top-1/4 right-0 h-[600px] w-[600px] rounded-full bg-accent/3 blur-3xl" />
          </div>
          <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="grid gap-12 lg:grid-cols-2 lg:gap-16">
              <div className="flex flex-col justify-center">
                <div className="mb-6 flex flex-wrap gap-2">
                  {heroBadges.map((badge) => (
                    <span
                      key={badge.label}
                      className="inline-flex items-center gap-1.5 rounded-full border border-border/50 bg-secondary/50 px-3 py-1 text-xs font-medium text-secondary-foreground"
                    >
                      <Icon name={badge.icon} className="h-3 w-3" />
                      {badge.label}
                    </span>
                  ))}
                </div>

                <h1 className="mb-6 text-balance text-4xl font-bold tracking-tight text-foreground sm:text-5xl lg:text-6xl">
                  Turn product context into working software.
                </h1>
                <p className="mb-8 max-w-xl text-pretty text-lg leading-relaxed text-muted-foreground">
                  Cleartify pulls together feedback, notes, prototypes, decisions, and code into one shared
                  context layer so humans and AI agents can plan, build, and ship faster.
                </p>

                <div className="flex flex-col gap-4 sm:flex-row">
                  <ButtonLink href={signupHref} className="group gap-2">
                    Start with a prototype
                    <Icon name="arrow-right" className="h-4 w-4 transition-transform group-hover:translate-x-1" />
                  </ButtonLink>
                  <ButtonLink href="#how-it-works" variant="secondary">
                    See how it works
                  </ButtonLink>
                </div>
              </div>

              <div className="relative">
                <div className="overflow-hidden rounded-2xl border border-border/50 bg-card/50 p-1 shadow-2xl backdrop-blur-sm">
                  <div className="rounded-xl bg-card">
                    <div className="flex items-center gap-2 border-b border-border/50 px-4 py-3">
                      <div className="flex gap-1.5">
                        <div className="h-3 w-3 rounded-full bg-red-500/80" />
                        <div className="h-3 w-3 rounded-full bg-yellow-500/80" />
                        <div className="h-3 w-3 rounded-full bg-green-500/80" />
                      </div>
                      <div className="flex-1 text-center">
                        <span className="text-xs text-muted-foreground">Cleartify - Context Workspace</span>
                      </div>
                    </div>
                    <div className="p-4">
                      <div className="mb-4 flex gap-1 rounded-lg bg-secondary/50 p-1">
                        <button className="rounded-md bg-card px-3 py-1.5 text-xs font-medium text-foreground shadow-sm">
                          Inbox
                        </button>
                        <button className="px-3 py-1.5 text-xs text-muted-foreground">Work Objects</button>
                        <button className="px-3 py-1.5 text-xs text-muted-foreground">Agents</button>
                      </div>
                      <div className="space-y-3">
                        <div className="rounded-lg border border-border/50 bg-secondary/30 p-3">
                          <div className="mb-2 flex items-center gap-2">
                            <div className="h-2 w-2 rounded-full bg-accent" />
                            <span className="text-xs font-medium text-foreground">User Feedback</span>
                            <span className="ml-auto text-xs text-muted-foreground">2m ago</span>
                          </div>
                          <p className="text-xs leading-relaxed text-muted-foreground">
                            Dashboard feels slow on mobile. Need to optimize chart rendering and load order.
                          </p>
                        </div>
                        <div className="rounded-lg border border-border/50 bg-secondary/30 p-3">
                          <div className="mb-2 flex items-center gap-2">
                            <div className="h-2 w-2 rounded-full bg-blue-500" />
                            <span className="text-xs font-medium text-foreground">Prototype Link</span>
                            <span className="ml-auto text-xs text-muted-foreground">15m ago</span>
                          </div>
                          <p className="text-xs leading-relaxed text-muted-foreground">
                            New onboarding flow mockup ready for review. Five screens attached.
                          </p>
                        </div>
                        <div className="rounded-lg border border-accent/30 bg-accent/5 p-3">
                          <div className="mb-2 flex items-center gap-2">
                            <Icon name="bot" className="h-3 w-3 text-accent" />
                            <span className="text-xs font-medium text-foreground">Agent Execution</span>
                            <span className="ml-auto rounded-full bg-accent/20 px-2 py-0.5 text-[10px] font-medium text-accent">
                              Running
                            </span>
                          </div>
                          <p className="text-xs leading-relaxed text-muted-foreground">
                            Generating implementation plan from mobile optimization context.
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                <div className="absolute -right-4 -bottom-4 h-32 w-32 rounded-full bg-accent/10 blur-2xl" />
              </div>
            </div>
          </div>
        </section>

        <section className="border-y border-border/50 bg-card/30 py-12">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <p className="mb-8 text-center text-sm font-medium text-muted-foreground">Used by fast-moving product teams</p>
            <div className="flex flex-wrap items-center justify-center gap-x-12 gap-y-6">
              {logos.map((logo) => (
                <div
                  key={logo}
                  className="text-lg font-semibold tracking-tight text-muted-foreground/60 transition-colors hover:text-muted-foreground"
                >
                  {logo}
                </div>
              ))}
            </div>
          </div>
        </section>

        <section className="py-24 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="mx-auto max-w-3xl text-center">
              <h2 className="mb-6 text-balance text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
                Old tools were built for handoffs.
              </h2>
              <p className="text-lg leading-relaxed text-muted-foreground">
                Traditional issue trackers were designed for a world where PMs scoped work, engineers picked it up
                later, and process filled the gaps. Today, AI can absorb more of the procedural work. The bottleneck
                is no longer tickets. It&apos;s context.
              </p>
            </div>

            <div className="mt-16 grid gap-8 md:grid-cols-2">
              <div className="rounded-2xl border border-border/50 bg-card/50 p-8">
                <div className="mb-6 flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-destructive/10">
                    <Icon name="x" className="h-5 w-5 text-destructive" />
                  </div>
                  <h3 className="text-lg font-semibold text-foreground">Old World</h3>
                </div>
                <ul className="space-y-4">
                  {oldWorldItems.map((item) => (
                    <li key={item} className="flex items-start gap-3">
                      <div className="mt-1 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-muted">
                        <Icon name="x" className="h-3 w-3 text-muted-foreground" />
                      </div>
                      <span className="text-muted-foreground">{item}</span>
                    </li>
                  ))}
                </ul>
              </div>

              <div className="rounded-2xl border border-accent/20 bg-accent/5 p-8">
                <div className="mb-6 flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-accent/20">
                    <Icon name="check" className="h-5 w-5 text-accent" />
                  </div>
                  <h3 className="text-lg font-semibold text-foreground">New World</h3>
                </div>
                <ul className="space-y-4">
                  {newWorldItems.map((item) => (
                    <li key={item} className="flex items-start gap-3">
                      <div className="mt-1 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-accent/20">
                        <Icon name="check" className="h-3 w-3 text-accent" />
                      </div>
                      <span className="text-foreground">{item}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </section>

        <section id="product" className="border-t border-border/50 bg-card/30 py-24 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="Everything agents need. Everything teams forget to keep connected."
              description="A complete system for turning context into execution."
            />

            <div className="mt-16 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {featureCards.map((card) => (
                <div
                  key={card.title}
                  className="group rounded-2xl border border-border/50 bg-card p-6 transition-all hover:border-accent/30 hover:bg-card/80"
                >
                  <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-secondary transition-colors group-hover:bg-accent/10">
                    <Icon name={card.icon} className="h-6 w-6 text-muted-foreground transition-colors group-hover:text-accent" />
                  </div>
                  <h3 className="mb-2 text-lg font-semibold text-foreground">{card.title}</h3>
                  <p className="text-sm leading-relaxed text-muted-foreground">{card.description}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section id="how-it-works" className="py-24 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="From messy inputs to shipped code."
              description="A streamlined workflow from idea to production."
            />

            <div className="relative mt-16">
              <div className="absolute top-24 left-0 right-0 hidden h-px bg-gradient-to-r from-transparent via-border to-transparent lg:block" />
              <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-4">
                {steps.map((step) => (
                  <div key={step.number} className="relative">
                    <div className="flex flex-col items-center text-center">
                      <div className="relative mb-6">
                        <div className="flex h-16 w-16 items-center justify-center rounded-2xl border border-border/50 bg-card shadow-lg">
                          <Icon name={step.icon} className="h-7 w-7 text-accent" />
                        </div>
                        <div className="absolute -top-2 -right-2 flex h-6 w-6 items-center justify-center rounded-full bg-accent text-xs font-bold text-accent-foreground">
                          {step.number}
                        </div>
                      </div>
                      <h3 className="mb-2 text-lg font-semibold text-foreground">{step.title}</h3>
                      <p className="text-sm leading-relaxed text-muted-foreground">{step.description}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        <section id="demo" className="border-t border-border/50 bg-card/30 py-24 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="See Cleartify in action"
              description="A unified workspace where context meets execution."
            />

            <div className="mt-12">
              <div className="mb-6 flex flex-wrap justify-center gap-2">
                {demoTabs.map((tab) => {
                  const selected = tab.id === activeDemoTab.id

                  return (
                    <button
                      key={tab.id}
                      type="button"
                      onClick={() => setActiveDemoId(tab.id)}
                      className={cn(
                        'flex items-center gap-2 rounded-full px-4 py-2 text-sm font-medium transition-all',
                        selected
                          ? 'bg-primary text-primary-foreground shadow-lg'
                          : 'bg-secondary/50 text-muted-foreground hover:bg-secondary hover:text-foreground',
                      )}
                    >
                      <Icon name={tab.icon} className="h-4 w-4" />
                      {tab.label}
                    </button>
                  )
                })}
              </div>

              <div className="overflow-hidden rounded-2xl border border-border/50 bg-card shadow-2xl">
                <div className="flex items-center justify-between border-b border-border/50 bg-secondary/30 px-6 py-3">
                  <div className="flex gap-1.5">
                    <div className="h-3 w-3 rounded-full bg-red-500/80" />
                    <div className="h-3 w-3 rounded-full bg-yellow-500/80" />
                    <div className="h-3 w-3 rounded-full bg-green-500/80" />
                  </div>
                  <span className="text-xs text-muted-foreground">app.cleartify.com</span>
                  <div className="w-16" />
                </div>

                <div className="min-h-[400px] p-6 lg:p-8">
                  <div className="space-y-4">
                    <div className="mb-6 flex items-center justify-between">
                      <h3 className="text-lg font-semibold text-foreground">{activeDemoTab.title}</h3>
                      <span className="rounded-full bg-accent/20 px-3 py-1 text-xs font-medium text-accent">
                        {activeDemoTab.badge}
                      </span>
                    </div>

                    <div className="space-y-3">
                      {activeDemoTab.items.map((item) => (
                        <div
                          key={`${activeDemoTab.id}-${item.title}`}
                          className={cn(
                            'flex items-center justify-between rounded-xl border p-4 transition-colors hover:bg-secondary/50',
                            item.tone === 'accent' && 'border-accent/30 bg-accent/5',
                            item.tone === 'info' && 'border-blue-500/20 bg-blue-500/5',
                            item.tone === 'warning' && 'border-yellow-500/20 bg-yellow-500/5',
                            item.tone === 'muted' && 'border-border/50 bg-secondary/20',
                          )}
                        >
                          <div className="flex items-center gap-4">
                            <div
                              className={cn(
                                'h-2 w-2 rounded-full',
                                item.tone === 'accent' && 'bg-accent',
                                item.tone === 'info' && 'bg-blue-400',
                                item.tone === 'warning' && 'bg-yellow-400',
                                item.tone === 'muted' && 'bg-muted-foreground/50',
                              )}
                            />
                            <div>
                              <p className="font-medium text-foreground">{item.title}</p>
                              <p className="text-sm text-muted-foreground">{item.meta}</p>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="py-24 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="Not another issue tracker."
              description="See how Cleartify compares to traditional tools and generic AI assistants."
            />

            <div className="mt-12 overflow-x-auto">
              <table className="w-full min-w-[600px]">
                <thead>
                  <tr className="border-b border-border/50">
                    <th className="pb-4 pr-4 text-left text-sm font-medium text-muted-foreground">Feature</th>
                    <th className="px-4 pb-4 text-center text-sm font-medium text-muted-foreground">
                      Traditional Issue Tracker
                    </th>
                    <th className="px-4 pb-4 text-center text-sm font-medium text-muted-foreground">
                      Generic AI Coding Tool
                    </th>
                    <th className="pb-4 pl-4 text-center">
                      <span className="rounded-full bg-accent/20 px-3 py-1 text-sm font-semibold text-accent">
                        Cleartify
                      </span>
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/50">
                  {matrixRows.map((row) => (
                    <tr key={row.feature}>
                      <td className="py-4 pr-4 text-sm font-medium text-foreground">{row.feature}</td>
                      <td className="px-4 py-4">
                        <MatrixCell value={row.traditional} />
                      </td>
                      <td className="px-4 py-4">
                        <MatrixCell value={row.genericAI} />
                      </td>
                      <td className="py-4 pl-4">
                        <MatrixCell value={row.cleartify} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <section id="agents" className="border-t border-border/50 bg-card/30 py-24 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="Built for teams that build fast."
              description="Whether you're a startup or scale-up, Cleartify adapts to how you work."
            />

            <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {audiences.map((audience) => (
                <div
                  key={audience.title}
                  className="group flex items-start gap-4 rounded-2xl border border-border/50 bg-card p-6 transition-all hover:border-accent/30 hover:bg-card/80"
                >
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-secondary transition-colors group-hover:bg-accent/10">
                    <Icon name={audience.icon} className="h-5 w-5 text-muted-foreground transition-colors group-hover:text-accent" />
                  </div>
                  <div>
                    <h3 className="mb-1 font-semibold text-foreground">{audience.title}</h3>
                    <p className="text-sm leading-relaxed text-muted-foreground">{audience.description}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section className="py-24 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="Trusted by product teams"
              description="Hear from teams shipping faster with Cleartify."
            />

            <div className="mt-12 grid gap-6 md:grid-cols-3">
              {testimonials.map((testimonial) => (
                <div key={testimonial.name} className="flex flex-col rounded-2xl border border-border/50 bg-card p-6">
                  <blockquote className="flex-1">
                    <p className="leading-relaxed text-foreground">"{testimonial.quote}"</p>
                  </blockquote>
                  <div className="mt-6 flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-secondary text-sm font-semibold text-foreground">
                      {testimonial.initials}
                    </div>
                    <div>
                      <p className="text-sm font-medium text-foreground">{testimonial.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {testimonial.role}, {testimonial.company}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section id="pricing" className="border-t border-border/50 bg-card/30 py-24 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="Simple pricing for teams that want to ship."
              description="Start free. Scale as you grow. No surprises."
            />

            <div className="mt-12 grid gap-6 lg:grid-cols-3">
              {pricingTiers.map((tier) => (
                <div
                  key={tier.name}
                  className={cn(
                    'relative flex flex-col rounded-2xl border p-8',
                    tier.featured
                      ? 'border-accent bg-accent/5 shadow-lg shadow-accent/10'
                      : 'border-border/50 bg-card',
                  )}
                >
                  {tier.featured ? (
                    <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                      <span className="rounded-full bg-accent px-3 py-1 text-xs font-medium text-accent-foreground">
                        Most Popular
                      </span>
                    </div>
                  ) : null}

                  <div className="mb-6">
                    <h3 className="text-lg font-semibold text-foreground">{tier.name}</h3>
                    <div className="mt-2 flex items-baseline gap-1">
                      <span className="text-4xl font-bold text-foreground">{tier.price}</span>
                      {tier.period ? <span className="text-sm text-muted-foreground">{tier.period}</span> : null}
                    </div>
                    <p className="mt-2 text-sm text-muted-foreground">{tier.description}</p>
                  </div>

                  <ul className="mb-8 flex-1 space-y-3">
                    {tier.features.map((feature) => (
                      <li key={`${tier.name}-${feature}`} className="flex items-start gap-3">
                        <Icon name="check" className="mt-0.5 h-4 w-4 shrink-0 text-accent" />
                        <span className="text-sm text-muted-foreground">{feature}</span>
                      </li>
                    ))}
                  </ul>

                  <ButtonLink
                    href={tier.ctaHref}
                    variant={tier.featured ? 'primary' : 'secondary'}
                    className="w-full justify-center"
                  >
                    {tier.ctaLabel}
                  </ButtonLink>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section id="faq" className="py-24 lg:py-32">
          <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="Frequently asked questions"
              description="Everything you need to know about Cleartify."
            />

            <div className="mt-12">
              {faqItems.map((item) => (
                <FAQEntry
                  key={item.question}
                  item={item}
                  open={openQuestion === item.question}
                  onToggle={() =>
                    setOpenQuestion((current) => (current === item.question ? '' : item.question))
                  }
                />
              ))}
            </div>
          </div>
        </section>

        <section className="relative overflow-hidden border-t border-border/50 bg-card/50 py-24 lg:py-32">
          <div className="pointer-events-none absolute inset-0">
            <div className="absolute bottom-0 left-1/2 h-[500px] w-[800px] -translate-x-1/2 translate-y-1/2 rounded-full bg-accent/5 blur-3xl" />
          </div>
          <div className="relative mx-auto max-w-4xl px-4 text-center sm:px-6 lg:px-8">
            <h2 className="mb-6 text-balance text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
              Build from context, not ceremony.
            </h2>
            <p className="mx-auto mb-10 max-w-2xl text-lg leading-relaxed text-muted-foreground">
              Start with a prototype, connect your product signals, and let Cleartify turn them into execution.
            </p>
            <div className="flex flex-col justify-center gap-4 sm:flex-row">
              <ButtonLink href={signupHref} className="group gap-2 px-8">
                Start with a prototype
                <Icon name="arrow-right" className="h-4 w-4 transition-transform group-hover:translate-x-1" />
              </ButtonLink>
              <ButtonLink href={demoHref} variant="secondary" className="px-8">
                Book a demo
              </ButtonLink>
            </div>
          </div>
        </section>

        <footer className="border-t border-border/50 bg-background">
          <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8 lg:py-16">
            <div className="grid gap-8 lg:grid-cols-6">
              <div className="lg:col-span-2">
                <a href="/" className="flex items-center gap-2">
                  <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
                    <span className="text-sm font-bold text-primary-foreground">C</span>
                  </div>
                  <span className="text-lg font-semibold tracking-tight text-foreground">Cleartify</span>
                </a>
                <p className="mt-4 max-w-xs text-sm leading-relaxed text-muted-foreground">
                  Turn product context into working software. From feedback and prototypes to shipped code, without
                  ticket theater.
                </p>
                <div className="mt-6 flex gap-4">
                  <a
                    href="#"
                    className="flex h-9 w-9 items-center justify-center rounded-full bg-secondary text-muted-foreground transition-colors hover:bg-secondary/80 hover:text-foreground"
                    aria-label="Twitter"
                  >
                    <Icon name="twitter" className="h-4 w-4" />
                  </a>
                  <a
                    href="#"
                    className="flex h-9 w-9 items-center justify-center rounded-full bg-secondary text-muted-foreground transition-colors hover:bg-secondary/80 hover:text-foreground"
                    aria-label="GitHub"
                  >
                    <Icon name="github" className="h-4 w-4" />
                  </a>
                  <a
                    href="#"
                    className="flex h-9 w-9 items-center justify-center rounded-full bg-secondary text-muted-foreground transition-colors hover:bg-secondary/80 hover:text-foreground"
                    aria-label="LinkedIn"
                  >
                    <Icon name="linkedin" className="h-4 w-4" />
                  </a>
                </div>
              </div>

              {footerColumns.map((column) => (
                <div key={column.title}>
                  <h3 className="mb-4 text-sm font-semibold text-foreground">{column.title}</h3>
                  <ul className="space-y-3">
                    {column.links.map((link) => (
                      <li key={`${column.title}-${link.label}`}>
                        <a href={link.href} className="text-sm text-muted-foreground transition-colors hover:text-foreground">
                          {link.label}
                        </a>
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>

            <div className="mt-12 border-t border-border/50 pt-8">
              <p className="text-center text-sm text-muted-foreground">(c) 2026 Cleartify. All rights reserved.</p>
            </div>
          </div>
        </footer>
      </main>
    </div>
  )
}
