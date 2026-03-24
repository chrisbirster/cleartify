import { useEffect, useState, type ReactNode, type SVGProps } from 'react'

type IconName =
  | 'activity'
  | 'alert-triangle'
  | 'arrow-right'
  | 'check-circle'
  | 'clipboard-list'
  | 'cloud'
  | 'database'
  | 'dollar-sign'
  | 'file-warning'
  | 'github'
  | 'git-branch'
  | 'info'
  | 'layers'
  | 'link'
  | 'linkedin'
  | 'map-pin'
  | 'menu'
  | 'message-square'
  | 'monitor'
  | 'network'
  | 'play'
  | 'search'
  | 'settings'
  | 'shield'
  | 'terminal'
  | 'trending-up'
  | 'user-x'
  | 'users'
  | 'wallet'
  | 'x'

type BadgeTone =
  | 'accent'
  | 'blue'
  | 'green'
  | 'neutral'
  | 'orange'
  | 'purple'
  | 'red'
  | 'slate'
  | 'teal'
  | 'yellow'

type NavLink = {
  label: string
  href: string
}

type SurfaceCard = {
  title: string
  description: string
  icon: IconName
}

type PainPoint = {
  title: string
  description: string
  icon: IconName
}

type AccessSurface = {
  name: string
  description: string
  badge: string
  tone: BadgeTone
}

type Step = {
  number: string
  title: string
  description: string
  icon: IconName
}

type WorkflowItem = {
  title: string
  description: string
  icon: IconName
  previewTitle: string
  previewItems: Array<{
    label: string
    meta: string
    tone: 'blue' | 'red' | 'yellow'
  }>
}

type Integration = {
  name: string
  description: string
  status: 'Available' | 'Coming Soon'
}

type AudienceCard = {
  title: string
  description: string
  icon: IconName
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
  (import.meta.env.DEV ? 'http://localhost:8000' : 'https://app.cleartify.com')
).replace(/\/$/, '')

const signInHref = `${appOrigin}/login`

const navLinks: NavLink[] = [
  {label: 'Product', href: '#product'},
  {label: 'Integrations', href: '#integrations'},
  {label: 'How It Works', href: '#how-it-works'},
  {label: 'Security', href: '#security'},
  {label: 'Contact', href: '#contact'},
]

const surfaceCards: SurfaceCard[] = [
  {
    title: 'Vendor Tools',
    description: 'Copilot, Cursor, ChatGPT',
    icon: 'layers',
  },
  {
    title: 'Cloud Accounts',
    description: 'Bedrock, Vertex AI',
    icon: 'cloud',
  },
  {
    title: 'API Projects',
    description: 'OpenAI, Anthropic',
    icon: 'git-branch',
  },
  {
    title: 'Internal Apps',
    description: 'RAG apps, Gateways',
    icon: 'terminal',
  },
]

const painPoints: PainPoint[] = [
  {
    title: 'AI adoption is outrunning governance',
    description:
      'Teams adopt new AI tools faster than policies can keep up. Shadow AI usage grows while IT scrambles to understand what is being used.',
    icon: 'trending-up',
  },
  {
    title: 'The same model appears through multiple access paths',
    description:
      'GPT-4 might be accessed via API, Copilot, a team workspace, and an internal wrapper, each with different governance implications.',
    icon: 'network',
  },
  {
    title: 'Ownership is fragmented or missing',
    description:
      'Service accounts, shared workspaces, and inherited roles make it unclear who is responsible for each AI system.',
    icon: 'user-x',
  },
  {
    title: 'Approvals are disconnected from real usage',
    description:
      'Approval workflows exist in spreadsheets while actual usage happens in completely different systems.',
    icon: 'file-warning',
  },
  {
    title: 'Budgets and spend are fragmented',
    description:
      'AI costs appear across multiple invoices, accounts, and cost centers with no unified view of total spend.',
    icon: 'wallet',
  },
  {
    title: 'Audit evidence is manual and scattered',
    description:
      'Preparing for audits means hunting through emails, Slack threads, and vendor consoles to piece together compliance evidence.',
    icon: 'clipboard-list',
  },
]

const accessSurfaces: AccessSurface[] = [
  {
    name: 'GitHub Copilot',
    badge: 'Organization',
    description: 'Org admin, seat assignments, policy settings',
    tone: 'purple',
  },
  {
    name: 'OpenAI API',
    badge: 'Project',
    description: 'API keys, usage limits, model access',
    tone: 'green',
  },
  {
    name: 'Amazon Bedrock',
    badge: 'Account + IAM Role',
    description: 'AWS account, IAM policies, model access',
    tone: 'orange',
  },
  {
    name: 'Claude Workspace',
    badge: 'Workspace',
    description: 'Workspace admin, member access, usage',
    tone: 'yellow',
  },
  {
    name: 'Internal RAG App',
    badge: 'Application',
    description: 'App owner, data sources, access control',
    tone: 'blue',
  },
  {
    name: 'Cursor',
    badge: 'Workspace',
    description: 'License seats, team settings, usage',
    tone: 'teal',
  },
]

const steps: Step[] = [
  {
    number: '01',
    title: 'Connect your AI systems',
    description:
      'Integrate with OpenAI, GitHub Copilot, Amazon Bedrock, and your internal AI apps. Setup takes minutes, not weeks.',
    icon: 'link',
  },
  {
    number: '02',
    title: 'Map usage to your org',
    description:
      'Automatically map AI systems to teams, cost centers, and owners. See who is using what across your organization.',
    icon: 'map-pin',
  },
  {
    number: '03',
    title: 'Create a governed system of record',
    description:
      'Build a single source of truth for all AI usage. Track approvals, budgets, and ownership in one place.',
    icon: 'database',
  },
  {
    number: '04',
    title: 'Find what needs attention',
    description:
      'Surface orphaned systems, pending approvals, budget overruns, and compliance gaps before they become problems.',
    icon: 'search',
  },
]

const keyQuestions = [
  'Which teams are using AI?',
  'Through which tools and access surfaces?',
  'Who owns each one?',
  'What is approved?',
  'What is it costing us?',
  'What needs cleanup right now?',
]

const workflowItems: WorkflowItem[] = [
  {
    title: 'Org Map',
    description: 'See which teams and departments are using AI across your organization',
    icon: 'users',
    previewTitle: 'Org Map',
    previewItems: [
      {label: 'Engineering owns 12 governed systems', meta: 'Departments', tone: 'blue'},
      {label: 'Data Science has 8 active access surfaces', meta: 'Teams', tone: 'blue'},
      {label: 'Product needs ownership mapping on 1 workspace', meta: 'Review', tone: 'yellow'},
      {label: 'Marketing has 3 approved systems', meta: 'Visibility', tone: 'blue'},
    ],
  },
  {
    title: 'AI Systems Registry',
    description: 'A complete inventory of every AI system, access surface, and integration',
    icon: 'layers',
    previewTitle: 'AI Systems Registry',
    previewItems: [
      {label: 'GitHub Copilot is approved for Platform Team', meta: 'Registry', tone: 'blue'},
      {label: 'OpenAI API project prod-ml is mapped to ML Team', meta: 'Registry', tone: 'blue'},
      {label: 'Amazon Bedrock is governed under Cloud Team', meta: 'Registry', tone: 'blue'},
      {label: 'Cursor workspace is still pending approval', meta: 'Registry', tone: 'yellow'},
    ],
  },
  {
    title: 'Usage Explorer',
    description: 'Track usage patterns, spend, and trends across all AI systems',
    icon: 'activity',
    previewTitle: 'Usage Explorer',
    previewItems: [
      {label: 'Monthly spend is trending up 12% across OpenAI', meta: 'Usage', tone: 'blue'},
      {label: 'Bedrock usage is concentrated in Cloud Team', meta: 'Usage', tone: 'blue'},
      {label: 'Copilot adoption expanded into Product this month', meta: 'Usage', tone: 'blue'},
      {label: 'Budget pressure flagged on one OpenAI project', meta: 'Usage', tone: 'yellow'},
    ],
  },
  {
    title: 'Approvals & Findings',
    description: 'Manage approvals and surface issues that need attention',
    icon: 'shield',
    previewTitle: 'Approvals & Findings',
    previewItems: [
      {label: '3 orphaned workspaces found', meta: 'Finding', tone: 'red'},
      {label: 'Cursor Enterprise pending approval', meta: 'Approval', tone: 'yellow'},
      {label: 'Budget threshold reached: OpenAI', meta: 'Alert', tone: 'yellow'},
      {label: 'New system detected: Claude Team', meta: 'Discovery', tone: 'blue'},
    ],
  },
]

const integrations: Integration[] = [
  {
    name: 'OpenAI API',
    description: 'Connect your OpenAI organization and projects',
    status: 'Available',
  },
  {
    name: 'GitHub Copilot',
    description: 'Integrate with your Copilot organization',
    status: 'Available',
  },
  {
    name: 'Amazon Bedrock',
    description: 'Connect your AWS accounts and IAM roles',
    status: 'Available',
  },
  {
    name: 'Internal AI Apps',
    description: 'Register your custom AI gateways and apps',
    status: 'Available',
  },
  {
    name: 'ChatGPT Enterprise',
    description: 'Coming soon: Team and Enterprise workspace integration',
    status: 'Coming Soon',
  },
  {
    name: 'Anthropic Direct',
    description: 'Coming soon: Direct Claude API integration',
    status: 'Coming Soon',
  },
]

const primaryAudiences: AudienceCard[] = [
  {
    title: 'IT',
    description: 'Gain visibility into AI adoption across the organization',
    icon: 'monitor',
  },
  {
    title: 'Platform Engineering',
    description: 'Build a governed foundation for AI infrastructure',
    icon: 'settings',
  },
  {
    title: 'Internal Developer Platforms',
    description: 'Enable teams with self-service AI access and governance',
    icon: 'shield',
  },
]

const secondaryStakeholders = [
  {
    title: 'Security',
    description: 'Assess and manage AI-related risks',
  },
  {
    title: 'Risk & Compliance',
    description: 'Streamline audit and evidence collection',
  },
  {
    title: 'Finance',
    description: 'Track and control AI spend',
  },
  {
    title: 'Architecture',
    description: 'Maintain standards and governance',
  },
]

const footerColumns: FooterColumn[] = [
  {
    title: 'Product',
    links: [
      {label: 'Features', href: '#product'},
      {label: 'Integrations', href: '#integrations'},
      {label: 'How It Works', href: '#how-it-works'},
      {label: 'Pricing', href: '#contact'},
    ],
  },
  {
    title: 'Company',
    links: [
      {label: 'About', href: '#contact'},
      {label: 'Blog', href: '#contact'},
      {label: 'Careers', href: '#contact'},
      {label: 'Contact', href: '#contact'},
    ],
  },
  {
    title: 'Resources',
    links: [
      {label: 'Documentation', href: '#contact'},
      {label: 'API Reference', href: '#contact'},
      {label: 'Security', href: '#security'},
      {label: 'Status', href: '#contact'},
    ],
  },
  {
    title: 'Legal',
    links: [
      {label: 'Privacy', href: '#contact'},
      {label: 'Terms', href: '#contact'},
      {label: 'Security Policy', href: '#security'},
    ],
  },
]

const dashboardTeams = [
  {name: 'Engineering', systems: '12 systems', tone: 'green'},
  {name: 'Data Science', systems: '8 systems', tone: 'green'},
  {name: 'Product', systems: '5 systems', tone: 'yellow'},
  {name: 'Marketing', systems: '3 systems', tone: 'green'},
] as const

const registryRows = [
  {name: 'GitHub Copilot', meta: 'Organization', owner: 'Platform Team', status: 'Approved', tone: 'green'},
  {name: 'OpenAI API', meta: 'Project: prod-ml', owner: 'ML Team', status: 'Approved', tone: 'green'},
  {name: 'Amazon Bedrock', meta: 'Account + IAM', owner: 'Cloud Team', status: 'Approved', tone: 'green'},
  {name: 'Cursor', meta: 'Workspace', owner: 'Engineering', status: 'Pending', tone: 'yellow'},
] as const

const findingItems = [
  {label: '3 orphaned workspaces', tone: 'red'},
  {label: '2 pending approvals', tone: 'yellow'},
  {label: 'Budget threshold reached', tone: 'blue'},
] as const

const toneClasses: Record<BadgeTone, string> = {
  accent: 'border-[rgba(22,211,228,0.28)] bg-[rgba(22,211,228,0.12)] text-primary',
  blue: 'border-[rgba(64,169,255,0.24)] bg-[rgba(64,169,255,0.12)] text-[var(--info)]',
  green: 'border-[rgba(42,214,143,0.24)] bg-[rgba(42,214,143,0.12)] text-[var(--success)]',
  neutral: 'border-border bg-[rgba(10,20,27,0.72)] text-foreground',
  orange: 'border-[rgba(255,166,92,0.24)] bg-[rgba(255,166,92,0.12)] text-[var(--warning)]',
  purple: 'border-[rgba(172,124,255,0.24)] bg-[rgba(172,124,255,0.12)] text-[rgb(195,162,255)]',
  red: 'border-[rgba(255,95,109,0.24)] bg-[rgba(255,95,109,0.12)] text-[var(--destructive)]',
  slate: 'border-border bg-[rgba(12,20,28,0.8)] text-muted-foreground',
  teal: 'border-[rgba(22,211,228,0.24)] bg-[rgba(22,211,228,0.12)] text-primary',
  yellow: 'border-[rgba(246,196,83,0.24)] bg-[rgba(246,196,83,0.12)] text-[var(--warning)]',
}

const panelClass =
  'rounded-3xl border border-border bg-[rgba(9,17,24,0.72)] backdrop-blur-md shadow-[0_24px_80px_rgba(0,0,0,0.35)]'

const subtlePanelClass =
  'rounded-2xl border border-border bg-[rgba(10,20,27,0.76)] shadow-[0_18px_55px_rgba(0,0,0,0.2)]'

function cn(...parts: Array<string | false | null | undefined>) {
  return parts.filter(Boolean).join(' ')
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
  variant?: 'primary' | 'secondary' | 'ghost'
  className?: string
  onClick?: () => void
}) {
  return (
    <a
      href={href}
      onClick={onClick}
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-xl px-5 py-3 text-sm font-medium transition duration-200',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/70 focus-visible:ring-offset-2 focus-visible:ring-offset-background',
        variant === 'primary' &&
          'bg-primary text-primary-foreground hover:-translate-y-0.5 hover:shadow-[0_12px_32px_rgba(22,211,228,0.28)]',
        variant === 'secondary' &&
          'border border-border bg-[rgba(8,14,20,0.82)] text-foreground hover:border-[rgba(150,177,189,0.28)] hover:bg-secondary',
        variant === 'ghost' &&
          'text-muted-foreground hover:bg-secondary hover:text-foreground',
        className,
      )}
    >
      {children}
    </a>
  )
}

function Badge({
  children,
  tone,
  className,
}: {
  children: ReactNode
  tone: BadgeTone
  className?: string
}) {
  return (
    <span
      className={cn(
        'inline-flex items-center justify-center rounded-full border px-2.5 py-1 text-[11px] font-medium tracking-[0.02em]',
        toneClasses[tone],
        className,
      )}
    >
      {children}
    </span>
  )
}

function SectionHeading({
  eyebrow,
  title,
  description,
  centered = true,
}: {
  eyebrow?: string
  title: string
  description?: string
  centered?: boolean
}) {
  return (
    <div className={cn('max-w-3xl', centered ? 'mx-auto text-center' : 'text-left')}>
      {eyebrow ? (
        <Badge tone="accent" className="mb-6">
          {eyebrow}
        </Badge>
      ) : null}
      <h2 className="text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">{title}</h2>
      {description ? (
        <p className={cn('mt-4 text-base leading-7 text-muted-foreground sm:text-lg', centered && 'mx-auto')}>
          {description}
        </p>
      ) : null}
    </div>
  )
}

function Icon({name, className, ...props}: {name: IconName; className?: string} & SVGProps<SVGSVGElement>) {
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
      {name === 'activity' ? (
        <>
          <path d="M22 12h-2.5a2 2 0 0 0-1.9 1.45l-2.4 8.1a.25.25 0 0 1-.48 0L9.24 2.2a.25.25 0 0 0-.48 0l-2.35 8.35A2 2 0 0 1 4.5 12H2" />
        </>
      ) : null}
      {name === 'alert-triangle' ? (
        <>
          <path d="m21.73 18-8-14a2 2 0 0 0-3.46 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3" />
          <path d="M12 9v4" />
          <path d="M12 17h.01" />
        </>
      ) : null}
      {name === 'arrow-right' ? (
        <>
          <path d="M5 12h14" />
          <path d="m12 5 7 7-7 7" />
        </>
      ) : null}
      {name === 'check-circle' ? (
        <>
          <circle cx="12" cy="12" r="10" />
          <path d="m9 12 2 2 4-4" />
        </>
      ) : null}
      {name === 'clipboard-list' ? (
        <>
          <rect x="8" y="2" width="8" height="4" rx="1" />
          <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" />
          <path d="M12 11h4" />
          <path d="M12 16h4" />
          <path d="M8 11h.01" />
          <path d="M8 16h.01" />
        </>
      ) : null}
      {name === 'cloud' ? (
        <>
          <path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z" />
        </>
      ) : null}
      {name === 'database' ? (
        <>
          <ellipse cx="12" cy="5" rx="9" ry="3" />
          <path d="M3 5v14c0 1.66 4.03 3 9 3s9-1.34 9-3V5" />
          <path d="M3 12c0 1.66 4.03 3 9 3s9-1.34 9-3" />
        </>
      ) : null}
      {name === 'dollar-sign' ? (
        <>
          <line x1="12" y1="2" x2="12" y2="22" />
          <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
        </>
      ) : null}
      {name === 'file-warning' ? (
        <>
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
          <path d="M14 2v6h6" />
          <path d="M12 11v4" />
          <path d="M12 18h.01" />
        </>
      ) : null}
      {name === 'github' ? (
        <>
          <path
            fill="currentColor"
            stroke="none"
            d="M12 2a10 10 0 0 0-3.16 19.49c.5.09.68-.22.68-.48v-1.69c-2.78.6-3.37-1.2-3.37-1.2-.45-1.15-1.1-1.46-1.1-1.46-.9-.62.07-.61.07-.61 1 .07 1.53 1.03 1.53 1.03.89 1.53 2.34 1.08 2.9.82.1-.64.35-1.08.64-1.33-2.22-.25-4.55-1.11-4.55-4.95 0-1.09.39-1.98 1.03-2.68-.1-.25-.45-1.27.1-2.65 0 0 .84-.27 2.75 1.02A9.6 9.6 0 0 1 12 6.84c.85 0 1.7.11 2.5.33 1.91-1.29 2.75-1.02 2.75-1.02.54 1.38.2 2.4.1 2.65.64.7 1.02 1.59 1.02 2.68 0 3.85-2.34 4.69-4.57 4.94.36.31.68.92.68 1.85V21c0 .27.18.58.69.48A10 10 0 0 0 12 2Z"
          />
        </>
      ) : null}
      {name === 'git-branch' ? (
        <>
          <path d="M15 6a3 3 0 1 0-3-3v11" />
          <path d="M6 18a3 3 0 1 0 3 3V8a3 3 0 0 1 3-3h3" />
          <circle cx="18" cy="6" r="3" />
          <circle cx="6" cy="18" r="3" />
        </>
      ) : null}
      {name === 'info' ? (
        <>
          <circle cx="12" cy="12" r="10" />
          <path d="M12 16v-4" />
          <path d="M12 8h.01" />
        </>
      ) : null}
      {name === 'layers' ? (
        <>
          <path d="M12 2 2 7l10 5 10-5-10-5Z" />
          <path d="m2 12 10 5 10-5" />
          <path d="m2 17 10 5 10-5" />
        </>
      ) : null}
      {name === 'link' ? (
        <>
          <path d="M9 17H7A5 5 0 0 1 7 7h2" />
          <path d="M15 7h2a5 5 0 1 1 0 10h-2" />
          <line x1="8" y1="12" x2="16" y2="12" />
        </>
      ) : null}
      {name === 'linkedin' ? (
        <>
          <path
            fill="currentColor"
            stroke="none"
            d="M19 3A2 2 0 0 1 21 5v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h14ZM8.34 9.56H5.79V18h2.55V9.56ZM7.07 5.5a1.48 1.48 0 1 0 0 2.96 1.48 1.48 0 0 0 0-2.96ZM18 13.06c0-2.55-1.36-3.74-3.18-3.74-1.47 0-2.12.81-2.49 1.37h-.03v-1.13H9.86c.03.75 0 8.44 0 8.44h2.54v-4.72c0-.25.02-.5.09-.68.2-.5.65-1.02 1.41-1.02 1 0 1.4.77 1.4 1.9V18H18v-4.94Z"
          />
        </>
      ) : null}
      {name === 'map-pin' ? (
        <>
          <path d="M20 10c0 5-5.54 10.2-7.4 11.8a1 1 0 0 1-1.2 0C9.54 20.2 4 15 4 10a8 8 0 0 1 16 0Z" />
          <circle cx="12" cy="10" r="3" />
        </>
      ) : null}
      {name === 'menu' ? (
        <>
          <path d="M4 6h16" />
          <path d="M4 12h16" />
          <path d="M4 18h16" />
        </>
      ) : null}
      {name === 'message-square' ? (
        <>
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
        </>
      ) : null}
      {name === 'monitor' ? (
        <>
          <rect x="2" y="3" width="20" height="14" rx="2" />
          <path d="M8 21h8" />
          <path d="M12 17v4" />
        </>
      ) : null}
      {name === 'network' ? (
        <>
          <rect x="16" y="16" width="6" height="6" rx="1" />
          <rect x="2" y="16" width="6" height="6" rx="1" />
          <rect x="9" y="2" width="6" height="6" rx="1" />
          <path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3" />
          <path d="M12 12V8" />
        </>
      ) : null}
      {name === 'play' ? (
        <>
          <path d="M6 4.5a2 2 0 0 1 3-1.72l10 5.8a2 2 0 0 1 0 3.44l-10 5.8A2 2 0 0 1 6 16.1Z" />
        </>
      ) : null}
      {name === 'search' ? (
        <>
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.35-4.35" />
        </>
      ) : null}
      {name === 'settings' ? (
        <>
          <path d="M12 3a2 2 0 0 1 2 2 2 2 0 0 0 2.82 1.82 2 2 0 0 1 2.6 2.6A2 2 0 0 0 21 12a2 2 0 0 1-2 2 2 2 0 0 0-1.82 2.82 2 2 0 0 1-2.6 2.6A2 2 0 0 0 12 21a2 2 0 0 1-2-2 2 2 0 0 0-2.82-1.82 2 2 0 0 1-2.6-2.6A2 2 0 0 0 3 12a2 2 0 0 1 2-2 2 2 0 0 0 1.82-2.82 2 2 0 0 1 2.6-2.6A2 2 0 0 0 12 3Z" />
          <circle cx="12" cy="12" r="3" />
        </>
      ) : null}
      {name === 'shield' ? (
        <>
          <path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.5 3.8 17 5 19 5a1 1 0 0 1 1 1z" />
        </>
      ) : null}
      {name === 'terminal' ? (
        <>
          <path d="m4 17 6-6-6-6" />
          <path d="M12 19h8" />
        </>
      ) : null}
      {name === 'trending-up' ? (
        <>
          <path d="M16 7h6v6" />
          <path d="m22 7-8.5 8.5-5-5L2 17" />
        </>
      ) : null}
      {name === 'user-x' ? (
        <>
          <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
          <circle cx="9" cy="7" r="4" />
          <path d="m17 8 5 5" />
          <path d="m22 8-5 5" />
        </>
      ) : null}
      {name === 'users' ? (
        <>
          <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
          <circle cx="9" cy="7" r="4" />
          <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
          <path d="M16 3.13a4 4 0 0 1 0 7.75" />
        </>
      ) : null}
      {name === 'wallet' ? (
        <>
          <path d="M19 7V4a1 1 0 0 0-1-1H5a2 2 0 1 0 0 4h15a1 1 0 0 1 1 1v4h-3a2 2 0 1 0 0 4h3v3a1 1 0 0 1-1 1H5a2 2 0 0 1-2-2V5" />
        </>
      ) : null}
      {name === 'x' ? (
        <>
          <path
            fill="currentColor"
            stroke="none"
            d="M18.9 3h2.9l-6.32 7.2L23 21h-5.9l-4.62-6.07L7.16 21H4.24l6.76-7.72L3 3h6.04l4.17 5.53L18.9 3Zm-1.03 16.24h1.61L8.18 4.67H6.45l11.42 14.57Z"
          />
        </>
      ) : null}
    </svg>
  )
}

function Header({
  scrolled,
  menuOpen,
  onToggleMenu,
  onCloseMenu,
}: {
  scrolled: boolean
  menuOpen: boolean
  onToggleMenu: () => void
  onCloseMenu: () => void
}) {
  return (
    <header
      className={cn(
        'fixed inset-x-0 top-0 z-50 transition-all duration-300',
        scrolled ? 'border-b border-border bg-[rgba(4,9,12,0.82)] backdrop-blur-xl' : 'bg-transparent',
      )}
    >
      <nav className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4 sm:px-6 lg:px-8 lg:py-5">
        <a href="/" className="flex items-center gap-3" onClick={onCloseMenu}>
          <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-[0_12px_32px_rgba(22,211,228,0.28)]">
            <Icon name="layers" className="h-5 w-5" />
          </div>
          <span className="text-lg font-semibold tracking-tight text-foreground">Cleartify</span>
        </a>

        <div className="hidden items-center gap-8 lg:flex">
          {navLinks.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="text-sm text-muted-foreground transition hover:text-foreground"
            >
              {link.label}
            </a>
          ))}
        </div>

        <div className="hidden items-center gap-3 lg:flex">
          <ButtonLink href={signInHref} variant="ghost" className="px-3 py-2">
            Sign in
          </ButtonLink>
          <ButtonLink href="#contact" className="px-4 py-2.5">
            Book a demo
          </ButtonLink>
        </div>

        <button
          type="button"
          aria-label={menuOpen ? 'Close navigation menu' : 'Open navigation menu'}
          aria-expanded={menuOpen}
          onClick={onToggleMenu}
          className="inline-flex h-11 w-11 items-center justify-center rounded-xl border border-border bg-[rgba(9,17,24,0.78)] text-muted-foreground transition hover:text-foreground lg:hidden"
        >
          <Icon name="menu" className="h-5 w-5" />
        </button>
      </nav>

      <div
        className={cn(
          'overflow-hidden border-t border-border bg-[rgba(4,9,12,0.96)] backdrop-blur-xl transition-[max-height,opacity] duration-300 lg:hidden',
          menuOpen ? 'max-h-96 opacity-100' : 'max-h-0 opacity-0',
        )}
      >
        <div className="mx-auto flex max-w-7xl flex-col gap-2 px-4 py-4 sm:px-6">
          {navLinks.map((link) => (
            <a
              key={link.href}
              href={link.href}
              onClick={onCloseMenu}
              className="rounded-xl px-3 py-3 text-sm text-muted-foreground transition hover:bg-secondary hover:text-foreground"
            >
              {link.label}
            </a>
          ))}
          <div className="mt-3 flex flex-col gap-3">
            <ButtonLink href={signInHref} variant="secondary" onClick={onCloseMenu}>
              Sign in
            </ButtonLink>
            <ButtonLink href="#contact" onClick={onCloseMenu}>
              Book a demo
            </ButtonLink>
          </div>
        </div>
      </div>
    </header>
  )
}

function HeroDashboard() {
  return (
    <div className="relative mx-auto mt-16 max-w-6xl lg:mt-24">
      <div className="absolute inset-4 rounded-[2rem] bg-[radial-gradient(circle_at_top,rgba(22,211,228,0.18),transparent_55%)] blur-3xl" />
      <div className={cn('relative overflow-hidden p-1', panelClass)}>
        <div className="flex items-center gap-2 border-b border-border px-4 py-3">
          <div className="flex gap-2">
            <span className="h-3 w-3 rounded-full bg-[rgba(255,95,109,0.9)]" />
            <span className="h-3 w-3 rounded-full bg-[rgba(246,196,83,0.9)]" />
            <span className="h-3 w-3 rounded-full bg-[rgba(42,214,143,0.9)]" />
          </div>
          <div className="flex flex-1 justify-center">
            <div className="rounded-full border border-border bg-[rgba(10,20,27,0.84)] px-4 py-1 text-xs text-muted-foreground">
              app.cleartify.com/dashboard
            </div>
          </div>
        </div>

        <div className="grid gap-4 p-4 lg:grid-cols-[1.05fr_1fr_0.9fr] lg:p-6">
          <div className={cn('p-4', subtlePanelClass)}>
            <div className="mb-4 flex items-center gap-2">
              <Icon name="users" className="h-4 w-4 text-primary" />
              <span className="text-sm font-medium text-foreground">Org Map</span>
            </div>
            <div className="space-y-3">
              {dashboardTeams.map((team) => (
                <div
                  key={team.name}
                  className="flex items-center justify-between rounded-xl border border-[rgba(132,164,186,0.12)] bg-[rgba(4,9,12,0.46)] px-3 py-2.5"
                >
                  <div className="flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full bg-primary/70" />
                    <span className="text-sm text-foreground">{team.name}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground">{team.systems}</span>
                    <Icon
                      name={team.tone === 'green' ? 'check-circle' : 'alert-triangle'}
                      className={cn('h-3.5 w-3.5', team.tone === 'green' ? 'text-[var(--success)]' : 'text-[var(--warning)]')}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className={cn('p-4', subtlePanelClass)}>
            <div className="mb-4 flex items-center gap-2">
              <Icon name="layers" className="h-4 w-4 text-primary" />
              <span className="text-sm font-medium text-foreground">AI Systems Registry</span>
            </div>
            <div className="space-y-3">
              {registryRows.map((row) => (
                <div
                  key={row.name}
                  className="rounded-xl border border-[rgba(132,164,186,0.12)] bg-[rgba(4,9,12,0.46)] p-3"
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-sm font-medium text-foreground">{row.name}</span>
                    <Badge tone={row.tone === 'green' ? 'green' : 'yellow'}>{row.status}</Badge>
                  </div>
                  <div className="mt-2 flex items-center justify-between text-xs text-muted-foreground">
                    <span>{row.meta}</span>
                    <span>{row.owner}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className={cn('p-4', subtlePanelClass)}>
                <div className="mb-3 flex items-center gap-2">
                  <Icon name="dollar-sign" className="h-4 w-4 text-primary" />
                  <span className="text-xs text-muted-foreground">Monthly Spend</span>
                </div>
                <span className="text-2xl font-semibold text-foreground">$47.2k</span>
              </div>
              <div className={cn('p-4', subtlePanelClass)}>
                <div className="mb-3 flex items-center gap-2">
                  <Icon name="shield" className="h-4 w-4 text-primary" />
                  <span className="text-xs text-muted-foreground">Compliance</span>
                </div>
                <span className="text-2xl font-semibold text-foreground">94%</span>
              </div>
            </div>

            <div className={cn('p-4', subtlePanelClass)}>
              <div className="mb-4 flex items-center gap-2">
                <Icon name="alert-triangle" className="h-4 w-4 text-[var(--warning)]" />
                <span className="text-sm font-medium text-foreground">Findings</span>
              </div>
              <div className="space-y-2">
                {findingItems.map((item) => (
                  <div
                    key={item.label}
                    className="flex items-center gap-2 rounded-xl border border-[rgba(132,164,186,0.12)] bg-[rgba(4,9,12,0.46)] px-3 py-2.5"
                  >
                    <span
                      className={cn(
                        'h-1.5 w-1.5 rounded-full',
                        item.tone === 'red' && 'bg-[var(--destructive)]',
                        item.tone === 'yellow' && 'bg-[var(--warning)]',
                        item.tone === 'blue' && 'bg-[var(--info)]',
                      )}
                    />
                    <span className="text-xs text-muted-foreground">{item.label}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function GovernanceWorkflow({
  activeIndex,
  onSelect,
}: {
  activeIndex: number
  onSelect: (index: number) => void
}) {
  const activeWorkflow = workflowItems[activeIndex]

  return (
    <section className="relative py-20 lg:py-32">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <SectionHeading
          title="Built for the workflows IT actually has"
          description="Purpose-built views for the work you need to do"
        />

        <div className="mt-16 grid gap-8 lg:grid-cols-[0.85fr_1.35fr]">
          <div className="space-y-3">
            {workflowItems.map((item, index) => (
              <button
                key={item.title}
                type="button"
                onClick={() => onSelect(index)}
                className={cn(
                  'flex w-full items-start gap-4 rounded-2xl border p-4 text-left transition duration-200',
                  index === activeIndex
                    ? 'border-[rgba(22,211,228,0.36)] bg-[rgba(9,17,24,0.96)] shadow-[0_18px_40px_rgba(0,0,0,0.22)]'
                    : 'border-border bg-[rgba(9,17,24,0.62)] hover:border-[rgba(150,177,189,0.26)] hover:bg-[rgba(10,20,27,0.82)]',
                )}
              >
                <div
                  className={cn(
                    'flex h-10 w-10 shrink-0 items-center justify-center rounded-xl',
                    index === activeIndex ? 'bg-primary/15 text-primary' : 'bg-secondary text-muted-foreground',
                  )}
                >
                  <Icon name={item.icon} className="h-5 w-5" />
                </div>
                <div>
                  <h3 className={cn('font-medium', index === activeIndex ? 'text-foreground' : 'text-muted-foreground')}>
                    {item.title}
                  </h3>
                  <p className="mt-1 text-sm leading-6 text-muted-foreground">{item.description}</p>
                </div>
              </button>
            ))}
          </div>

          <div className={cn('p-6', panelClass)}>
            <div className="mb-6 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/15 text-primary">
                <Icon name={activeWorkflow.icon} className="h-5 w-5" />
              </div>
              <span className="font-medium text-foreground">{activeWorkflow.previewTitle}</span>
            </div>
            <div className="space-y-3">
              {activeWorkflow.previewItems.map((item) => (
                <div
                  key={item.label}
                  className="rounded-xl border border-[rgba(132,164,186,0.12)] bg-[rgba(10,20,27,0.72)] px-4 py-3"
                >
                  <div className="flex items-start gap-3">
                    <span
                      className={cn(
                        'mt-1.5 h-2 w-2 rounded-full',
                        item.tone === 'red' && 'bg-[var(--destructive)]',
                        item.tone === 'yellow' && 'bg-[var(--warning)]',
                        item.tone === 'blue' && 'bg-[var(--info)]',
                      )}
                    />
                    <div>
                      <p className="text-sm font-medium text-foreground">{item.label}</p>
                      <p className="mt-1 text-xs text-muted-foreground">{item.meta}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

function Footer() {
  return (
    <footer className="border-t border-border bg-[rgba(6,12,17,0.86)]">
      <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8 lg:py-16">
        <div className="grid gap-8 md:grid-cols-5">
          <div className="md:col-span-1">
            <a href="/" className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-foreground">
                <Icon name="layers" className="h-5 w-5" />
              </div>
              <span className="text-lg font-semibold text-foreground">Cleartify</span>
            </a>
            <p className="mt-4 max-w-xs text-sm leading-6 text-muted-foreground">
              AI governance system of record for IT and platform teams.
            </p>
          </div>

          {footerColumns.map((column) => (
            <div key={column.title}>
              <h4 className="text-sm font-medium text-foreground">{column.title}</h4>
              <ul className="mt-4 space-y-3">
                {column.links.map((link) => (
                  <li key={`${column.title}-${link.label}`}>
                    <a href={link.href} className="text-sm text-muted-foreground transition hover:text-foreground">
                      {link.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="mt-12 flex flex-col gap-4 border-t border-border pt-8 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-muted-foreground">© 2026 Cleartify, Inc. All rights reserved.</p>
          <div className="flex items-center gap-5 text-muted-foreground">
            <a href="#contact" aria-label="Cleartify on X" className="transition hover:text-foreground">
              <Icon name="x" className="h-5 w-5" />
            </a>
            <a href="#contact" aria-label="Cleartify on LinkedIn" className="transition hover:text-foreground">
              <Icon name="linkedin" className="h-5 w-5" />
            </a>
            <a href="#contact" aria-label="Cleartify on GitHub" className="transition hover:text-foreground">
              <Icon name="github" className="h-5 w-5" />
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}

export function App() {
  const [menuOpen, setMenuOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)
  const [activeWorkflow, setActiveWorkflow] = useState(3)

  useEffect(() => {
    const updateScrollState = () => {
      setScrolled(window.scrollY > 16)
    }

    updateScrollState()
    window.addEventListener('scroll', updateScrollState, {passive: true})
    return () => window.removeEventListener('scroll', updateScrollState)
  }, [])

  const closeMenu = () => {
    setMenuOpen(false)
  }

  return (
    <div className="marketing-shell min-h-screen bg-background text-foreground">
      <Header
        scrolled={scrolled}
        menuOpen={menuOpen}
        onToggleMenu={() => setMenuOpen((value) => !value)}
        onCloseMenu={closeMenu}
      />

      <main>
        <section className="relative overflow-hidden px-4 pb-20 pt-32 sm:px-6 lg:px-8 lg:pb-32 lg:pt-40">
          <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(22,211,228,0.12),transparent_38%)]" />
          <div className="pointer-events-none absolute left-1/2 top-0 h-[34rem] w-[34rem] -translate-x-1/2 rounded-full bg-[rgba(22,211,228,0.08)] blur-3xl" />

          <div className="relative mx-auto max-w-7xl">
            <div className="mx-auto max-w-4xl text-center">
              <div className="inline-flex items-center gap-3 rounded-full border border-border bg-[rgba(10,20,27,0.72)] px-4 py-2 text-sm text-muted-foreground">
                <span className="h-2 w-2 rounded-full bg-primary shadow-[0_0_18px_rgba(22,211,228,0.9)]" />
                Built for IT and platform teams managing AI across vendor tools, cloud accounts, and internal AI apps
              </div>

              <h1 className="mt-8 text-4xl font-semibold leading-tight tracking-tight text-foreground sm:text-5xl lg:text-7xl">
                Turn scattered AI usage into
                <br />
                <span className="bg-[linear-gradient(90deg,#16d3e4_0%,#4ef0c0_100%)] bg-clip-text text-transparent">
                  a governed system of record
                </span>
              </h1>

              <p className="mx-auto mt-6 max-w-2xl text-lg leading-8 text-muted-foreground sm:text-xl">
                Know which teams use AI, through which tools, under which owners, approvals, and budgets.
              </p>

              <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
                <ButtonLink href="#contact">
                  Book a demo
                  <Icon name="arrow-right" className="h-4 w-4" />
                </ButtonLink>
                <ButtonLink href="#how-it-works" variant="secondary">
                  <Icon name="play" className="h-4 w-4" />
                  See how it works
                </ButtonLink>
              </div>
            </div>

            <HeroDashboard />
          </div>
        </section>

        <section className="relative py-20 lg:py-28">
          <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,transparent,rgba(10,20,27,0.48),transparent)]" />
          <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="AI usage spreads faster than governance"
              description="Copilot, Cursor, Bedrock, OpenAI projects, internal RAG apps, team workspaces, service accounts, and cloud roles all create different governance boundaries. Cleartify gives you one place to see what exists, who owns it, what is approved, and what it costs."
            />

            <div className="mt-16 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              {surfaceCards.map((card) => (
                <div key={card.title} className={cn('p-6 transition duration-300 hover:-translate-y-1', subtlePanelClass)}>
                  <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
                    <Icon name={card.icon} className="h-6 w-6" />
                  </div>
                  <h3 className="text-lg font-medium text-foreground">{card.title}</h3>
                  <p className="mt-2 text-sm text-muted-foreground">{card.description}</p>
                </div>
              ))}
            </div>

            <div className="mt-12 flex flex-col items-center gap-4">
              <div className="flex items-center gap-3">
                <span className="h-px w-16 bg-[linear-gradient(90deg,transparent,rgba(132,164,186,0.4),transparent)]" />
                <span className="h-3 w-3 rounded-full bg-primary/70 shadow-[0_0_18px_rgba(22,211,228,0.85)]" />
                <span className="h-px w-16 bg-[linear-gradient(90deg,transparent,rgba(132,164,186,0.4),transparent)]" />
              </div>
              <Badge tone="accent" className="px-6 py-2 text-sm">
                One unified system of record
              </Badge>
            </div>
          </div>
        </section>

        <section className="py-20 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="Why AI governance gets messy fast"
              description="These are the real challenges IT and platform teams face every day."
            />

            <div className="mt-16 grid gap-6 md:grid-cols-2 xl:grid-cols-3">
              {painPoints.map((painPoint) => (
                <div
                  key={painPoint.title}
                  className="rounded-2xl border border-border bg-[rgba(9,17,24,0.56)] p-6 transition duration-300 hover:border-[rgba(150,177,189,0.26)] hover:bg-[rgba(10,20,27,0.78)]"
                >
                  <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-[rgba(255,95,109,0.08)] text-[var(--destructive)]">
                    <Icon name={painPoint.icon} className="h-6 w-6" />
                  </div>
                  <h3 className="text-lg font-medium leading-7 text-foreground">{painPoint.title}</h3>
                  <p className="mt-3 text-sm leading-7 text-muted-foreground">{painPoint.description}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section id="product" className="scroll-mt-24 py-20 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="grid items-start gap-12 lg:grid-cols-[1.05fr_0.95fr] lg:gap-16">
              <div>
                <SectionHeading
                  eyebrow="Core Concept"
                  title="The real problem is not just which model"
                  centered={false}
                />
                <p className="mt-4 text-xl font-medium text-primary">It is how AI is being accessed and governed.</p>
                <div className="mt-6 space-y-4 text-muted-foreground">
                  <p className="leading-7">
                    An <span className="font-medium text-foreground">access surface</span> is the specific route through
                    which AI is accessed and governed. The same model family can appear through multiple access surfaces,
                    each with different owners, logs, budgets, and policies.
                  </p>
                  <p className="leading-7">
                    Understanding access surfaces is the foundation of effective AI governance. Without this clarity, you
                    are governing blindly.
                  </p>
                </div>

                <div className={cn('mt-8 p-6', subtlePanelClass)}>
                  <div className="mb-4 flex items-center gap-3">
                    <div className="rounded-xl bg-secondary px-3 py-1.5 text-sm font-medium text-foreground">GPT-4</div>
                    <span className="text-sm text-muted-foreground">accessed through</span>
                  </div>
                  <div className="grid gap-2 sm:grid-cols-2">
                    {['OpenAI API', 'GitHub Copilot', 'Internal Gateway', 'Team Workspace'].map((item) => (
                      <div
                        key={item}
                        className="flex items-center gap-2 rounded-xl border border-[rgba(132,164,186,0.12)] bg-[rgba(4,9,12,0.42)] px-3 py-2 text-sm text-muted-foreground"
                      >
                        <span className="h-1.5 w-1.5 rounded-full bg-primary" />
                        {item}
                      </div>
                    ))}
                  </div>
                  <p className="mt-4 text-xs text-muted-foreground">
                    Each surface has different owners, budgets, and governance boundaries
                  </p>
                </div>
              </div>

              <div>
                <h3 className="mb-6 text-lg font-medium text-foreground">Example Access Surfaces</h3>
                <div className="space-y-4">
                  {accessSurfaces.map((surface) => (
                    <div
                      key={surface.name}
                      className="group flex items-start justify-between gap-4 rounded-2xl border border-border bg-[rgba(9,17,24,0.56)] p-4 transition duration-300 hover:border-[rgba(22,211,228,0.26)] hover:bg-[rgba(10,20,27,0.8)]"
                    >
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-3">
                          <h4 className="font-medium text-foreground">{surface.name}</h4>
                          <Badge tone={surface.tone}>{surface.badge}</Badge>
                        </div>
                        <p className="mt-2 text-sm text-muted-foreground">{surface.description}</p>
                      </div>
                      <span className="mt-2 h-2 w-2 shrink-0 rounded-full bg-primary/70 transition group-hover:bg-primary" />
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </section>

        <section id="how-it-works" className="scroll-mt-24 py-20 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading title="How Cleartify works" description="Four steps to operational AI governance" />

            <div className="mt-16 grid gap-6 md:grid-cols-2 xl:grid-cols-4">
              {steps.map((step, index) => (
                <div key={step.number} className="relative">
                  {index < steps.length - 1 ? (
                    <div className="absolute left-full top-12 hidden h-px w-full bg-[linear-gradient(90deg,rgba(132,164,186,0.12),rgba(22,211,228,0.28),rgba(132,164,186,0.12))] xl:block" />
                  ) : null}
                  <div className="relative h-full rounded-2xl border border-border bg-[rgba(9,17,24,0.58)] p-6">
                    <div className="mb-4 flex items-center gap-4">
                      <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
                        <Icon name={step.icon} className="h-6 w-6" />
                      </div>
                      <span className="text-sm font-medium text-primary">{step.number}</span>
                    </div>
                    <h3 className="text-lg font-medium text-foreground">{step.title}</h3>
                    <p className="mt-3 text-sm leading-7 text-muted-foreground">{step.description}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section className="py-20 lg:py-24">
          <div className="mx-auto grid max-w-7xl gap-10 px-4 sm:px-6 lg:grid-cols-[0.95fr_1.05fr] lg:px-8">
            <div>
              <SectionHeading
                title="What Cleartify helps you answer"
                description="These are the questions that keep IT and platform leaders up at night. Cleartify gives you the answers."
                centered={false}
              />
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              {keyQuestions.map((question) => (
                <div
                  key={question}
                  className="flex items-start gap-3 rounded-2xl border border-border bg-[rgba(9,17,24,0.56)] px-4 py-4"
                >
                  <Icon name="check-circle" className="mt-0.5 h-5 w-5 shrink-0 text-primary" />
                  <span className="text-sm leading-6 text-foreground">{question}</span>
                </div>
              ))}
            </div>
          </div>
        </section>

        <GovernanceWorkflow activeIndex={activeWorkflow} onSelect={setActiveWorkflow} />

        <section id="integrations" className="scroll-mt-24 relative py-20 lg:py-32">
          <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,transparent,rgba(10,20,27,0.34),transparent)]" />
          <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="Start with the systems you already use"
              description="Cleartify complements vendor admin consoles. It does not replace them."
            />

            <div className="mt-16 grid gap-6 md:grid-cols-2 xl:grid-cols-3">
              {integrations.map((integration) => (
                <div
                  key={integration.name}
                  className="rounded-2xl border border-border bg-[rgba(9,17,24,0.56)] p-6 transition duration-300 hover:border-[rgba(22,211,228,0.24)] hover:bg-[rgba(10,20,27,0.8)]"
                >
                  <div className="mb-5 flex items-start justify-between gap-4">
                    <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-secondary">
                      <div className="h-6 w-6 rounded-md bg-primary/35" />
                    </div>
                    <Badge tone={integration.status === 'Available' ? 'green' : 'slate'}>{integration.status}</Badge>
                  </div>
                  <h3 className="text-lg font-medium text-foreground">{integration.name}</h3>
                  <p className="mt-2 text-sm leading-7 text-muted-foreground">{integration.description}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section className="py-20 lg:py-32">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <SectionHeading
              title="Designed for IT and platform teams"
              description="Built by platform engineers, for platform engineers"
            />

            <div className="mt-16 grid gap-8 lg:grid-cols-3">
              {primaryAudiences.map((audience) => (
                <div key={audience.title} className={cn('p-8 transition duration-300 hover:border-[rgba(22,211,228,0.24)]', subtlePanelClass)}>
                  <div className="mb-6 flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                    <Icon name={audience.icon} className="h-7 w-7" />
                  </div>
                  <h3 className="text-xl font-medium text-foreground">{audience.title}</h3>
                  <p className="mt-3 text-sm leading-7 text-muted-foreground">{audience.description}</p>
                </div>
              ))}
            </div>

            <div className={cn('mt-10 p-6 lg:p-8', subtlePanelClass)}>
              <h3 className="text-sm font-medium uppercase tracking-[0.18em] text-muted-foreground">Also valuable for</h3>
              <div className="mt-5 grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
                {secondaryStakeholders.map((stakeholder) => (
                  <div key={stakeholder.title} className="flex items-start gap-3">
                    <span className="mt-2 h-2 w-2 rounded-full bg-primary" />
                    <div>
                      <p className="font-medium text-foreground">{stakeholder.title}</p>
                      <p className="mt-1 text-sm text-muted-foreground">{stakeholder.description}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        <section id="security" className="scroll-mt-24 py-20 lg:py-32">
          <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
            <SectionHeading eyebrow="Honest by Design" title="We are clear about what Cleartify does" />

            <div className={cn('mt-12 p-8', panelClass)}>
              <div className="flex items-start gap-4">
                <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                  <Icon name="info" className="h-5 w-5" />
                </div>
                <div>
                  <h3 className="text-lg font-medium text-foreground">
                    Cleartify v1 helps you govern approved AI usage across known systems.
                  </h3>
                  <p className="mt-2 leading-7 text-muted-foreground">
                    We focus on doing this exceptionally well, rather than making broad promises we cannot keep.
                  </p>
                </div>
              </div>

              <div className="mt-8 border-t border-border pt-6">
                <h4 className="text-sm font-medium uppercase tracking-[0.16em] text-muted-foreground">
                  What we do not promise:
                </h4>
                <div className="mt-4 grid gap-3 sm:grid-cols-2">
                  {[
                    'Visibility into every unmanaged consumer account',
                    'Prompt-level inspection across all tools',
                    'Endpoint enforcement in v1',
                    'Full legal automation',
                  ].map((item) => (
                    <div
                      key={item}
                      className="flex items-start gap-3 rounded-xl border border-[rgba(132,164,186,0.12)] bg-[rgba(10,20,27,0.72)] px-4 py-3"
                    >
                      <Icon name="x" className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">{item}</span>
                    </div>
                  ))}
                </div>
              </div>

              <p className="mt-6 text-center text-sm text-muted-foreground">
                This honesty is a feature, not a bug. It is why enterprises trust us.
              </p>
            </div>
          </div>
        </section>

        <section id="contact" className="scroll-mt-24 relative overflow-hidden px-4 py-20 sm:px-6 lg:px-8 lg:py-32">
          <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(10,20,27,0.06),rgba(22,211,228,0.08))]" />
          <div className="pointer-events-none absolute inset-x-0 bottom-0 h-48 bg-[radial-gradient(circle_at_center,rgba(22,211,228,0.14),transparent_60%)]" />
          <div className="relative mx-auto max-w-4xl text-center">
            <h2 className="text-4xl font-semibold tracking-tight text-foreground sm:text-5xl">
              Bring order to AI usage before it becomes a larger governance problem
            </h2>
            <p className="mx-auto mt-6 max-w-2xl text-lg leading-8 text-muted-foreground">
              Track which teams use AI, through which access surfaces, under which owners, approvals, and budgets.
            </p>
            <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
              <ButtonLink href="#contact">
                Book a demo
                <Icon name="arrow-right" className="h-4 w-4" />
              </ButtonLink>
              <ButtonLink href="#contact" variant="secondary">
                <Icon name="message-square" className="h-4 w-4" />
                Talk to us about your stack
              </ButtonLink>
            </div>
          </div>
        </section>
      </main>

      <Footer />
    </div>
  )
}
