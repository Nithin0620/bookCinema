import { useState } from "react";

const items = [
  {
    q: "Which tools does Solace connect to?",
    a: "Solace works with Outlook, Gmail, Slack, Zoom, Figma and Linear, and pulls them into one place so nothing slips between apps.",
  },
  {
    q: "How does Solace decide what is a priority?",
    a: "It weighs deadlines, who is asking and how often a thread comes back, then ranks your day so the one thing that can’t wait sits at the top.",
  },
  {
    q: "Is my data and chat history secure?",
    a: "Everything is encrypted in transit and at rest, nothing is used to train models, and you can export or delete your history at any time.",
  },
  {
    q: "Can I try Solace before I pay?",
    a: "Yes. Every workspace starts with a 14-day trial with every feature switched on, and no card needed.",
  },
];

export function Faq() {
  const [open, setOpen] = useState(0);

  return (
    <section className="faq-section">
      <h2 className="faq-title">Questions, answered</h2>
      <div className="faq">
        {items.map((item, index) => (
          <div className="q" data-open={open === index ? "" : undefined} key={item.q}>
            <button
              type="button"
              aria-expanded={open === index}
              onClick={() => setOpen(open === index ? -1 : index)}
            >
              {item.q}
              <svg viewBox="0 0 14 14" aria-hidden="true">
                <path d="M3 5l4 4 4-4" />
              </svg>
            </button>
            <div className="a">
              <div>
                <p>{item.a}</p>
              </div>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
