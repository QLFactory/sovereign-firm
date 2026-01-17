"use client";

import { useState, useEffect } from "react";
import Link from "next/link";

// Code lines for typewriter effect - defined outside component to avoid re-creation
const CODE_LINES = [
  '// AI Agents building your app...',
  'const app = await sovereignFirm.create({',
  '  name: "TaskFlow",',
  '  stack: { frontend: "React", backend: "Node.js" },',
  '  features: ["auth", "dashboard", "api"],',
  '});',
  '',
  '// Generating 47 production-ready files...',
  '✓ Frontend: 12 React components',
  '✓ Backend: 13 API endpoints',
  '✓ Database: Prisma schema + migrations',
  '✓ DevOps: Docker, K8s, Helm, CI/CD',
  '✓ Infrastructure: Terraform modules',
  '✓ Monitoring: Prometheus + Grafana',
  '',
  '// Project delivered in 2 hours, not 2 months',
];

// Animated code typing effect
function TypewriterCode() {
  const [displayedLines, setDisplayedLines] = useState<string[]>([]);
  const [currentLine, setCurrentLine] = useState(0);
  const [currentChar, setCurrentChar] = useState(0);

  useEffect(() => {
    if (currentLine >= CODE_LINES.length) {
      // Reset after a pause
      const timeout = setTimeout(() => {
        setDisplayedLines([]);
        setCurrentLine(0);
        setCurrentChar(0);
      }, 3000);
      return () => clearTimeout(timeout);
    }

    const line = CODE_LINES[currentLine];

    if (currentChar < line.length) {
      const timeout = setTimeout(() => {
        setDisplayedLines(prev => {
          // Ensure array has all previous lines filled
          const newLines = [...prev];
          while (newLines.length < currentLine) {
            newLines.push(CODE_LINES[newLines.length] || '');
          }
          newLines[currentLine] = line.substring(0, currentChar + 1);
          return newLines;
        });
        setCurrentChar(c => c + 1);
      }, 25);
      return () => clearTimeout(timeout);
    } else {
      const timeout = setTimeout(() => {
        setCurrentLine(l => l + 1);
        setCurrentChar(0);
      }, 100);
      return () => clearTimeout(timeout);
    }
  }, [currentLine, currentChar]);

  return (
    <div className="font-mono text-sm leading-relaxed">
      {displayedLines.map((line, i) => {
        // Guard against undefined lines
        const text = line ?? '';
        return (
          <div
            key={i}
            className={`${
              text.startsWith('//') ? 'text-[var(--silver)]' :
              text.startsWith('✓') ? 'text-[var(--emerald-glow)]' :
              text.includes(':') && !text.includes('//') ? 'text-[var(--cyan-glow)]' :
              'text-[var(--pearl)]'
            }`}
          >
            {text}
            {i === displayedLines.length - 1 && (
              <span className="inline-block w-2 h-4 bg-[var(--cyan-glow)] ml-1 animate-pulse" />
            )}
          </div>
        );
      })}
    </div>
  );
}

// Floating particles - uses seeded positions to avoid hydration mismatch
function Particles() {
  // Pre-computed particle positions to avoid SSR/client mismatch
  const particles = [
    { left: 15, top: 20, delay: 0, duration: 8 },
    { left: 85, top: 15, delay: 1.2, duration: 12 },
    { left: 25, top: 70, delay: 0.5, duration: 9 },
    { left: 70, top: 85, delay: 2.1, duration: 11 },
    { left: 45, top: 30, delay: 0.8, duration: 7 },
    { left: 90, top: 60, delay: 1.5, duration: 10 },
    { left: 10, top: 45, delay: 2.8, duration: 13 },
    { left: 60, top: 10, delay: 0.3, duration: 8 },
    { left: 35, top: 90, delay: 1.9, duration: 14 },
    { left: 80, top: 40, delay: 3.2, duration: 9 },
    { left: 20, top: 55, delay: 0.7, duration: 11 },
    { left: 55, top: 75, delay: 2.4, duration: 12 },
    { left: 95, top: 25, delay: 1.1, duration: 8 },
    { left: 40, top: 65, delay: 3.5, duration: 10 },
    { left: 75, top: 5, delay: 0.9, duration: 13 },
    { left: 5, top: 80, delay: 2.6, duration: 9 },
    { left: 50, top: 50, delay: 1.8, duration: 11 },
    { left: 30, top: 35, delay: 3.0, duration: 7 },
    { left: 65, top: 95, delay: 0.4, duration: 12 },
    { left: 88, top: 70, delay: 2.2, duration: 10 },
  ];

  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      {particles.map((p, i) => (
        <div
          key={i}
          className="absolute w-1 h-1 bg-[var(--cyan-glow)] rounded-full opacity-40 animate-float"
          style={{
            left: `${p.left}%`,
            top: `${p.top}%`,
            animationDelay: `${p.delay}s`,
            animationDuration: `${p.duration}s`,
          }}
        />
      ))}
    </div>
  );
}

// Stats counter animation
function AnimatedNumber({ value, suffix = "" }: { value: number; suffix?: string }) {
  const [count, setCount] = useState(0);

  useEffect(() => {
    const duration = 2000;
    const steps = 60;
    const increment = value / steps;
    let current = 0;

    const timer = setInterval(() => {
      current += increment;
      if (current >= value) {
        setCount(value);
        clearInterval(timer);
      } else {
        setCount(Math.floor(current));
      }
    }, duration / steps);

    return () => clearInterval(timer);
  }, [value]);

  return <span>{count}{suffix}</span>;
}

// Navigation
function Navigation() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 20);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <nav className={`nav ${scrolled ? 'scrolled' : ''}`}>
      <div className="nav-inner">
        <Link href="/" className="nav-logo">
          <div className="nav-logo-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M12 2L2 7l10 5 10-5-10-5z" />
              <path d="M2 17l10 5 10-5" />
              <path d="M2 12l10 5 10-5" />
            </svg>
          </div>
          <span>Sovereign Firm</span>
        </Link>

        <div className="nav-links">
          <a href="#features" className="nav-link">Features</a>
          <a href="#how-it-works" className="nav-link">How It Works</a>
          <a href="#pricing" className="nav-link">Pricing</a>
          <Link href="/dashboard" className="btn btn-primary">
            Launch Console
          </Link>
        </div>
      </div>
    </nav>
  );
}

// Hero Section
function HeroSection() {
  return (
    <section className="relative min-h-screen flex items-center justify-center overflow-hidden">
      {/* Gradient mesh background */}
      <div className="gradient-mesh">
        <div className="gradient-orb gradient-orb-1" />
        <div className="gradient-orb gradient-orb-2" />
        <div className="gradient-orb gradient-orb-3" />
        <div className="grid-overlay" />
        <div className="noise-overlay" />
      </div>

      <Particles />

      <div className="container relative z-10 pt-32 pb-20">
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          {/* Left: Copy */}
          <div className="space-y-8">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass-subtle animate-slide-down">
              <span className="w-2 h-2 bg-[var(--emerald-glow)] rounded-full animate-pulse" />
              <span className="text-sm text-[var(--pearl)]">Now delivering enterprise projects</span>
            </div>

            <h1 className="text-display-xl text-gradient-white animate-slide-up">
              Ship in Hours,<br />
              <span className="text-gradient-cyan">Not Months</span>
            </h1>

            <p className="text-body-lg text-[var(--silver)] max-w-xl animate-slide-up stagger-2">
              Sovereign Firm replaces your entire development team with AI agents.
              Describe your app, and watch as our agents architect, code, test, and
              deploy production-ready software in hours.
            </p>

            <div className="flex flex-wrap gap-4 animate-slide-up stagger-3">
              <Link href="/dashboard" className="btn btn-primary text-base px-8 py-4">
                Start Building Free
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M5 12h14M12 5l7 7-7 7" />
                </svg>
              </Link>
              <a href="#how-it-works" className="btn btn-secondary text-base px-8 py-4">
                See How It Works
              </a>
            </div>

            {/* Social proof */}
            <div className="flex items-center gap-8 pt-4 animate-slide-up stagger-4">
              <div className="flex -space-x-3">
                {[...Array(5)].map((_, i) => (
                  <div
                    key={i}
                    className="w-10 h-10 rounded-full bg-gradient-to-br from-[var(--steel)] to-[var(--slate)] border-2 border-[var(--obsidian)] flex items-center justify-center text-xs font-bold"
                  >
                    {String.fromCharCode(65 + i)}
                  </div>
                ))}
              </div>
              <div className="text-sm">
                <div className="font-semibold text-[var(--ivory)]">500+ projects delivered</div>
                <div className="text-[var(--silver)]">Trusted by startups & enterprises</div>
              </div>
            </div>
          </div>

          {/* Right: Live code visualization */}
          <div className="relative animate-scale-in stagger-2">
            <div className="absolute -inset-4 bg-gradient-to-r from-[var(--cyan-glow)] via-[var(--violet-glow)] to-[var(--amber-glow)] rounded-2xl opacity-20 blur-2xl animate-gradient" />
            <div className="relative glass-card rounded-2xl overflow-hidden">
              {/* Terminal header */}
              <div className="flex items-center gap-2 px-4 py-3 bg-[var(--void)] border-b border-[var(--steel)]">
                <div className="flex gap-2">
                  <div className="w-3 h-3 rounded-full bg-[var(--rose-glow)]" />
                  <div className="w-3 h-3 rounded-full bg-[var(--amber-glow)]" />
                  <div className="w-3 h-3 rounded-full bg-[var(--emerald-glow)]" />
                </div>
                <div className="flex-1 text-center text-xs text-[var(--silver)]">sovereign-firm — agents working</div>
              </div>
              {/* Terminal content */}
              <div className="p-6 bg-[var(--void)]/80 min-h-[400px]">
                <TypewriterCode />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Scroll indicator */}
      <div className="absolute bottom-8 left-1/2 -translate-x-1/2 animate-float">
        <div className="w-6 h-10 rounded-full border-2 border-[var(--steel)] flex items-start justify-center p-2">
          <div className="w-1 h-2 bg-[var(--cyan-glow)] rounded-full animate-pulse" />
        </div>
      </div>
    </section>
  );
}

// Stats Section
function StatsSection() {
  const stats = [
    { value: 100, suffix: "x", label: "Faster Delivery" },
    { value: 90, suffix: "%", label: "Cost Reduction" },
    { value: 47, suffix: "+", label: "Files Per Project" },
    { value: 24, suffix: "/7", label: "Agent Availability" },
  ];

  return (
    <section className="py-20 relative">
      <div className="container">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
          {stats.map((stat, i) => (
            <div key={i} className="text-center">
              <div className="text-display-lg text-gradient-cyan mb-2">
                <AnimatedNumber value={stat.value} suffix={stat.suffix} />
              </div>
              <div className="text-[var(--silver)] text-sm">{stat.label}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// Features Section
function FeaturesSection() {
  const features = [
    {
      icon: (
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="3" y="3" width="18" height="18" rx="2" />
          <path d="M9 9h6M9 13h6M9 17h4" />
        </svg>
      ),
      title: "Full-Stack Generation",
      description: "Frontend, backend, database, and infrastructure. Complete applications with a single prompt.",
      color: "cyan",
    },
    {
      icon: (
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83" />
        </svg>
      ),
      title: "AI Agent Orchestra",
      description: "PM, Architect, Developer, QA, DevOps, and SRE agents collaborate to deliver your project.",
      color: "violet",
    },
    {
      icon: (
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
          <polyline points="22 4 12 14.01 9 11.01" />
        </svg>
      ),
      title: "Production Quality",
      description: "Type-safe code, comprehensive tests, security best practices, and CI/CD pipelines included.",
      color: "emerald",
    },
    {
      icon: (
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
          <polyline points="7.5 4.21 12 6.81 16.5 4.21" />
          <polyline points="7.5 19.79 7.5 14.6 3 12" />
          <polyline points="21 12 16.5 14.6 16.5 19.79" />
          <polyline points="3.27 6.96 12 12.01 20.73 6.96" />
          <line x1="12" y1="22.08" x2="12" y2="12" />
        </svg>
      ),
      title: "Cloud Native",
      description: "Kubernetes manifests, Helm charts, Terraform modules. Deploy anywhere with confidence.",
      color: "amber",
    },
    {
      icon: (
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" />
          <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" />
        </svg>
      ),
      title: "Self-Documenting",
      description: "API documentation, architecture diagrams, and runbooks generated automatically.",
      color: "rose",
    },
    {
      icon: (
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
        </svg>
      ),
      title: "Enterprise Security",
      description: "OWASP compliance, secrets management, security scanning, and audit logging built in.",
      color: "cyan",
    },
  ];

  const colorClasses: Record<string, string> = {
    cyan: "text-[var(--cyan-glow)] bg-[rgba(0,240,255,0.1)]",
    violet: "text-[var(--violet-glow)] bg-[rgba(168,85,247,0.1)]",
    emerald: "text-[var(--emerald-glow)] bg-[rgba(16,185,129,0.1)]",
    amber: "text-[var(--amber-glow)] bg-[rgba(255,184,0,0.1)]",
    rose: "text-[var(--rose-glow)] bg-[rgba(244,63,94,0.1)]",
  };

  return (
    <section id="features" className="section">
      <div className="container">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <div className="badge badge-cyan mb-4">Capabilities</div>
          <h2 className="text-display-md text-gradient-white mb-6">
            Everything You Need to Ship
          </h2>
          <p className="text-body-lg text-[var(--silver)]">
            Our AI agents handle every aspect of software development, from initial architecture
            to production deployment and ongoing operations.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {features.map((feature, i) => (
            <div key={i} className="card card-glow group">
              <div className={`w-12 h-12 rounded-xl ${colorClasses[feature.color]} flex items-center justify-center mb-4 transition-transform group-hover:scale-110`}>
                {feature.icon}
              </div>
              <h3 className="text-lg font-semibold text-[var(--ivory)] mb-2">{feature.title}</h3>
              <p className="text-[var(--silver)] text-sm">{feature.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// How It Works Section
function HowItWorksSection() {
  const steps = [
    {
      number: "01",
      title: "Describe Your Vision",
      description: "Tell our PM Agent what you want to build. It will ask clarifying questions and gather requirements.",
      visual: "💬",
    },
    {
      number: "02",
      title: "Watch Agents Work",
      description: "Our AI agents architect, code, test, and configure your application in real-time.",
      visual: "⚡",
    },
    {
      number: "03",
      title: "Review & Deploy",
      description: "Review generated artifacts, provide feedback, and deploy to any cloud with one click.",
      visual: "🚀",
    },
  ];

  return (
    <section id="how-it-works" className="section bg-[var(--carbon)]">
      <div className="container">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <div className="badge badge-amber mb-4">Process</div>
          <h2 className="text-display-md text-gradient-white mb-6">
            Three Steps to Production
          </h2>
          <p className="text-body-lg text-[var(--silver)]">
            From idea to deployed application in hours, not months. Our agents handle the complexity.
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-8">
          {steps.map((step, i) => (
            <div key={i} className="relative">
              {/* Connector line */}
              {i < steps.length - 1 && (
                <div className="hidden md:block absolute top-16 left-[60%] w-[80%] h-px bg-gradient-to-r from-[var(--steel)] to-transparent" />
              )}

              <div className="text-center">
                <div className="relative inline-block mb-6">
                  <div className="w-32 h-32 rounded-full glass-card flex items-center justify-center text-5xl">
                    {step.visual}
                  </div>
                  <div className="absolute -top-2 -right-2 w-10 h-10 rounded-full bg-gradient-to-br from-[var(--cyan-glow)] to-[var(--violet-glow)] flex items-center justify-center text-sm font-bold text-[var(--void)]">
                    {step.number}
                  </div>
                </div>
                <h3 className="text-xl font-semibold text-[var(--ivory)] mb-3">{step.title}</h3>
                <p className="text-[var(--silver)] text-sm max-w-xs mx-auto">{step.description}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// Pricing Section
function PricingSection() {
  const plans = [
    {
      name: "Starter",
      description: "For individual developers and small projects",
      price: "Free",
      period: "",
      features: [
        "5 projects / month",
        "Frontend generation",
        "Basic backend",
        "Community support",
      ],
      cta: "Get Started",
      highlighted: false,
    },
    {
      name: "Pro",
      description: "For teams shipping production applications",
      price: "$99",
      period: "/month",
      features: [
        "Unlimited projects",
        "Full-stack generation",
        "DevOps & Infrastructure",
        "SRE & Monitoring",
        "Priority support",
        "Custom tech stacks",
      ],
      cta: "Start Pro Trial",
      highlighted: true,
    },
    {
      name: "Enterprise",
      description: "For organizations with custom requirements",
      price: "Custom",
      period: "",
      features: [
        "Everything in Pro",
        "Self-hosted option",
        "SSO / SAML",
        "Custom agents",
        "SLA guarantee",
        "Dedicated support",
      ],
      cta: "Contact Sales",
      highlighted: false,
    },
  ];

  return (
    <section id="pricing" className="section">
      <div className="container">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <div className="badge badge-emerald mb-4">Pricing</div>
          <h2 className="text-display-md text-gradient-white mb-6">
            Start Free, Scale as You Grow
          </h2>
          <p className="text-body-lg text-[var(--silver)]">
            From side projects to enterprise deployments. Pay only for what you use.
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
          {plans.map((plan, i) => (
            <div
              key={i}
              className={`relative rounded-2xl p-8 ${
                plan.highlighted
                  ? 'bg-gradient-to-b from-[var(--cyan-glow)]/10 to-transparent border-2 border-[var(--cyan-glow)]/30'
                  : 'glass-card'
              }`}
            >
              {plan.highlighted && (
                <div className="absolute -top-4 left-1/2 -translate-x-1/2 badge badge-cyan">
                  Most Popular
                </div>
              )}

              <div className="mb-6">
                <h3 className="text-xl font-semibold text-[var(--ivory)] mb-2">{plan.name}</h3>
                <p className="text-sm text-[var(--silver)]">{plan.description}</p>
              </div>

              <div className="mb-6">
                <span className="text-4xl font-bold text-[var(--ivory)]">{plan.price}</span>
                <span className="text-[var(--silver)]">{plan.period}</span>
              </div>

              <ul className="space-y-3 mb-8">
                {plan.features.map((feature, j) => (
                  <li key={j} className="flex items-center gap-3 text-sm text-[var(--pearl)]">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--emerald-glow)" strokeWidth="2">
                      <polyline points="20 6 9 17 4 12" />
                    </svg>
                    {feature}
                  </li>
                ))}
              </ul>

              <Link
                href="/dashboard"
                className={`btn w-full ${plan.highlighted ? 'btn-primary' : 'btn-secondary'}`}
              >
                {plan.cta}
              </Link>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// CTA Section
function CTASection() {
  return (
    <section className="section">
      <div className="container">
        <div className="relative rounded-3xl overflow-hidden">
          {/* Background */}
          <div className="absolute inset-0 bg-gradient-to-br from-[var(--cyan-glow)]/20 via-[var(--violet-glow)]/20 to-[var(--amber-glow)]/20" />
          <div className="absolute inset-0 glass" />

          <div className="relative p-12 md:p-20 text-center">
            <h2 className="text-display-md text-[var(--ivory)] mb-6">
              Ready to Transform How You Build Software?
            </h2>
            <p className="text-body-lg text-[var(--pearl)] max-w-2xl mx-auto mb-8">
              Join hundreds of teams shipping faster with AI-powered development.
              Start free, no credit card required.
            </p>
            <div className="flex flex-wrap justify-center gap-4">
              <Link href="/dashboard" className="btn btn-primary text-base px-8 py-4">
                Launch Console
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M5 12h14M12 5l7 7-7 7" />
                </svg>
              </Link>
              <a href="#" className="btn btn-secondary text-base px-8 py-4">
                Schedule Demo
              </a>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

// Footer
function Footer() {
  return (
    <footer className="py-12 border-t border-[var(--steel)]">
      <div className="container">
        <div className="flex flex-col md:flex-row justify-between items-center gap-6">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-gradient-to-br from-[var(--cyan-glow)] to-[var(--violet-glow)] rounded-lg flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <path d="M12 2L2 7l10 5 10-5-10-5z" />
                <path d="M2 17l10 5 10-5" />
                <path d="M2 12l10 5 10-5" />
              </svg>
            </div>
            <span className="font-semibold text-[var(--ivory)]">Sovereign Firm</span>
          </div>

          <div className="flex gap-8 text-sm text-[var(--silver)]">
            <a href="#" className="hover:text-[var(--ivory)] transition-colors">Documentation</a>
            <a href="#" className="hover:text-[var(--ivory)] transition-colors">API</a>
            <a href="#" className="hover:text-[var(--ivory)] transition-colors">Status</a>
            <a href="#" className="hover:text-[var(--ivory)] transition-colors">Privacy</a>
            <a href="#" className="hover:text-[var(--ivory)] transition-colors">Terms</a>
          </div>

          <div className="text-sm text-[var(--silver)]">
            © 2024 Sovereign Firm. All rights reserved.
          </div>
        </div>
      </div>
    </footer>
  );
}

// Main Landing Page
export default function LandingPage() {
  return (
    <main className="min-h-screen bg-[var(--obsidian)]">
      <Navigation />
      <HeroSection />
      <StatsSection />
      <FeaturesSection />
      <HowItWorksSection />
      <PricingSection />
      <CTASection />
      <Footer />
    </main>
  );
}
