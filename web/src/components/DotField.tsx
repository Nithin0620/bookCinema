import { useEffect, useRef } from "react";

const C0 = [226, 106, 24];
const C1 = [248, 150, 64];
const C2 = [255, 214, 170];

const hash = (i: number, j: number) => {
  const h = Math.sin(i * 127.1 + j * 311.7) * 43758.5453;
  return h - Math.floor(h);
};

const mix = (a: number[], b: number[], t: number) =>
  a.map((v, k) => Math.round(v + (b[k] - v) * t));

export function DotField() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    const ctx = canvas?.getContext("2d");
    if (!canvas || !ctx) return;

    const calm = matchMedia("(prefers-reduced-motion: reduce)").matches;
    let W = 0;
    let H = 0;
    let dpr = 1;
    let step = 0;
    let cols = 0;
    let rows = 0;
    let seed: [number, number, number][] = [];
    let mx = -9999;
    let my = -9999;
    let tmx = -9999;
    let tmy = -9999;
    let frame = 0;

    const size = () => {
      dpr = Math.min(devicePixelRatio || 1, 2);
      W = canvas.clientWidth;
      H = canvas.clientHeight;
      canvas.width = W * dpr;
      canvas.height = H * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      step = Math.max(14, Math.min(W / 1600, H / 940) * 20.5);
      cols = Math.ceil(W / step) + 1;
      rows = Math.ceil(H / step) + 1;
      seed = [];
      for (let j = 0; j < rows; j++) {
        for (let i = 0; i < cols; i++) {
          seed.push([hash(i, j), hash(i + 71, j + 13), hash(i + 9, j + 97)]);
        }
      }
    };

    const onPointerMove = (event: PointerEvent) => {
      const rect = canvas.getBoundingClientRect();
      tmx = event.clientX - rect.left;
      tmy = event.clientY - rect.top;
    };
    const onPointerLeave = () => {
      tmx = -9999;
      tmy = -9999;
    };

    const draw = (t: number) => {
      ctx.clearRect(0, 0, W, H);
      mx += (tmx - mx) * 0.12;
      my += (tmy - my) * 0.12;
      const top = H * 0.1;
      const full = H * 0.36;
      const R = step * 0.37;

      for (let j = 0; j < rows; j++) {
        const y = j * step + step * 0.5;
        if (y < top - step * 2) continue;
        const v = Math.min(1, Math.max(0, (y - top) / (full - top)));
        const col =
          v < 0.55
            ? mix(C0, C1, v / 0.55)
            : mix(
                C1,
                C2,
                ((v - 0.55) / 0.45) * 0.85 +
                  Math.max(0, (y - H * 0.9) / (H * 0.1)) * 0.15,
              );
        ctx.fillStyle = `rgb(${col[0]},${col[1]},${col[2]})`;

        for (let i = 0; i < cols; i++) {
          const x = i * step + step * 0.5;
          const [a, b, c] = seed[j * cols + i];
          const wave = calm
            ? 0.7
            : 0.5 +
              0.5 *
                Math.sin(
                  t * (1.6 + b * 2.6) +
                    a * 6.283 +
                    Math.sin(t * 0.7 + i * 0.21 + j * 0.13) * 1.4,
                );
          const keep = Math.min(1, v * 1.6 + 0.12) - a * 0.6;
          if (keep <= 0 && c > 0.08) continue;
          let s =
            (0.42 + 0.58 * Math.pow(wave, 1.3)) *
            Math.min(1, 0.5 + v * 0.6) *
            (keep > 0 ? 1 : 0.8);
          const dx = x - mx;
          const dy = y - my;
          const d2 = dx * dx + dy * dy;
          const hov = Math.exp(-d2 / (step * step * 22));
          s = Math.min(1.05, s + hov * 0.7);
          const r = R * s;
          if (r < 0.6) continue;
          ctx.beginPath();
          ctx.arc(x, y, r, 0, 6.2832);
          ctx.fill();
        }
      }
    };

    const loop = (ms: number) => {
      draw(ms / 1000);
      frame = requestAnimationFrame(loop);
    };

    size();
    const observer = new ResizeObserver(size);
    observer.observe(canvas);
    addEventListener("resize", size);
    addEventListener("pointermove", onPointerMove);
    document.addEventListener("pointerleave", onPointerLeave);
    frame = requestAnimationFrame(loop);

    return () => {
      cancelAnimationFrame(frame);
      observer.disconnect();
      removeEventListener("resize", size);
      removeEventListener("pointermove", onPointerMove);
      document.removeEventListener("pointerleave", onPointerLeave);
    };
  }, []);

  return <canvas className="dot-field" ref={canvasRef} aria-hidden="true" />;
}
