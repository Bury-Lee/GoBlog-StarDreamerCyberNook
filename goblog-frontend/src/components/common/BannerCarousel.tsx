import { useState, useEffect } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import type { Banner } from '../../types';

interface BannerCarouselProps {
  banners: Banner[];
}

export default function BannerCarousel({ banners }: BannerCarouselProps) {
  const [current, setCurrent] = useState(0);

  useEffect(() => {
    if (banners.length <= 1) return;

    const timer = setInterval(() => {
      setCurrent((prev) => (prev + 1) % banners.length);
    }, 5000);

    return () => clearInterval(timer);
  }, [banners.length]);

  const prev = () => {
    setCurrent((prev) => (prev - 1 + banners.length) % banners.length);
  };

  const next = () => {
    setCurrent((prev) => (prev + 1) % banners.length);
  };

  if (!banners.length) return null;

  const currentBanner = banners[current];

  return (
    <div className="relative w-full h-[300px] sm:h-[350px] lg:h-[400px] rounded-2xl overflow-hidden group shadow-lg">
      <div className="absolute inset-0">
        <img
          src={currentBanner.image}
          alt={currentBanner.title}
          className="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
        />
        <div className="absolute inset-0 bg-gradient-to-t from-slate-900/80 via-slate-900/30 to-transparent" />
      </div>

      <div className="absolute bottom-0 left-0 right-0 p-6 sm:p-8">
        <h2 className="text-2xl sm:text-3xl lg:text-3xl font-bold text-white mb-2 leading-tight">
          {currentBanner.title}
        </h2>
        {currentBanner.link && (
          <a
            href={currentBanner.link}
            className="inline-flex items-center gap-2 text-blue-400 hover:text-blue-300 transition-colors text-sm sm:text-base"
          >
            了解更多 →
          </a>
        )}
      </div>

      {banners.length > 1 && (
        <>
          <button
            onClick={prev}
            className="absolute left-3 sm:left-4 top-1/2 -translate-y-1/2 w-9 h-9 sm:w-10 sm:h-10 rounded-full bg-white/90 backdrop-blur-sm flex items-center justify-center text-slate-800 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-white shadow-lg"
          >
            <ChevronLeft size={20} />
          </button>
          <button
            onClick={next}
            className="absolute right-3 sm:right-4 top-1/2 -translate-y-1/2 w-9 h-9 sm:w-10 sm:h-10 rounded-full bg-white/90 backdrop-blur-sm flex items-center justify-center text-slate-800 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-white shadow-lg"
          >
            <ChevronRight size={20} />
          </button>

          <div className="absolute bottom-3 sm:bottom-4 right-6 sm:right-8 flex gap-2">
            {banners.map((_, index) => (
              <button
                key={index}
                onClick={() => setCurrent(index)}
                className={`w-2 h-2 sm:w-2.5 sm:h-2.5 rounded-full transition-all ${
                  index === current
                    ? 'w-6 bg-blue-500'
                    : 'bg-white/60 hover:bg-white'
                }`}
              />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
