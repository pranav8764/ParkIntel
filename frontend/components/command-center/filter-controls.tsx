"use client";

// ============================================================================
// Filter Controls — Hour, date, station, and risk level selectors.
// ============================================================================

import { useState, useEffect, useRef } from "react";
import { useCommandStore } from "@/store/command-store";
import { formatHour } from "@/lib/format";
import { cn } from "@/lib/utils";
import { Clock, Building2, AlertTriangle, RotateCcw } from "lucide-react";

const STATIONS = [
  "ADUGODI",
  "ASHOK NAGAR",
  "BANASHANKARI",
  "BANASWADI",
  "BASAVANAGUDI",
  "BELLANDUR",
  "BYATARAYANAPURA",
  "CHAMARAJPET",
  "CHIKKABANAVARA",
  "CHIKKAJALA",
  "CITY MARKET",
  "CUBBON PARK",
  "DEVANAHALLI AIRPORT",
  "ELECTRONIC CITY",
  "HAL OLD AIRPORT",
  "HALASUR",
  "HALASURU GATE",
  "HEBBALA",
  "HENNURU",
  "HIGH GROUND",
  "HSR LAYOUT",
  "HULIMAVU",
  "J.P. NAGAR",
  "JALAHALLI",
  "JAYANAGARA",
  "JEEVANBHEEMANAGAR",
  "JNANABHARATHI",
  "K.G. HALLI",
  "K.R. PURA",
  "K.S. LAYOUT",
  "KAMAKSHIPALYA",
  "KENGERI",
  "KODIGEHALLI",
  "MADIWALA",
  "MAGADI ROAD",
  "MAHADEVAPURA",
  "MALLESHWARAM",
  "MICO LAYOUT",
  "NO POLICE STATION",
  "PEENYA",
  "PULIKESHINAGAR(F.TOWN)",
  "R.T. NAGAR",
  "RAJAJINAGAR",
  "SADASHIVANAGAR",
  "SHESHADRIPURAM",
  "SHIVAJINAGAR",
  "THALAGATTAPURA",
  "UPPARPET",
  "V.V.PURAM (C.PET)",
  "VIJAYANAGARA",
  "WHITEFIELD",
  "WILSON GARDEN",
  "YELAHANKA",
  "YESHWANTHPURA"
];

export function FilterControls() {
  const selectedHour = useCommandStore((s) => s.selectedHour);
  const setSelectedHour = useCommandStore((s) => s.setSelectedHour);
  const selectedPoliceStation = useCommandStore((s) => s.selectedPoliceStation);
  const setSelectedPoliceStation = useCommandStore(
    (s) => s.setSelectedPoliceStation
  );
  const selectedRiskLevel = useCommandStore((s) => s.selectedRiskLevel);
  const setSelectedRiskLevel = useCommandStore((s) => s.setSelectedRiskLevel);
  const resetFilters = useCommandStore((s) => s.resetFilters);

  const [inputValue, setInputValue] = useState(selectedPoliceStation ?? "");
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Sync input value with the global store (e.g. when reset is clicked)
  useEffect(() => {
    setInputValue(selectedPoliceStation ?? "");
  }, [selectedPoliceStation]);

  // Click outside detection to close suggestions
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  const hours = Array.from({ length: 24 }, (_, i) => i);

  const riskLevels = [
    { value: "", label: "All Risks" },
    { value: "HIGH", label: "High" },
    { value: "MEDIUM", label: "Medium" },
    { value: "LOW", label: "Low" },
  ];

  const suggestions = isOpen && inputValue.trim().length > 0
    ? STATIONS.filter((station) =>
        station.toLowerCase().includes(inputValue.toLowerCase())
      )
    : [];

  const handleSelect = (station: string) => {
    setSelectedPoliceStation(station);
    setInputValue(station);
    setIsOpen(false);
  };

  return (
    <div className="flex items-center gap-2 flex-wrap">
      {/* Hour Selector */}
      <div className="flex items-center gap-1.5">
        <Clock className="h-3.5 w-3.5 text-muted" />
        <select
          value={selectedHour}
          onChange={(e) => setSelectedHour(Number(e.target.value))}
          className={cn(
            "h-8 rounded-md px-2 text-xs font-mono",
            "bg-white/5 text-foreground border border-white/10",
            "transition-all duration-200",
            "hover:border-white/15 focus:border-gold/40 focus:outline-none focus:ring-1 focus:ring-gold/20",
            "cursor-pointer appearance-none"
          )}
        >
          {hours.map((h) => (
            <option key={h} value={h} className="bg-surface text-foreground">
              {formatHour(h)}
            </option>
          ))}
        </select>
      </div>

      {/* Risk Level */}
      <div className="flex items-center gap-1.5">
        <AlertTriangle className="h-3.5 w-3.5 text-muted" />
        <select
          value={selectedRiskLevel ?? ""}
          onChange={(e) =>
            setSelectedRiskLevel(e.target.value || null)
          }
          className={cn(
            "h-8 rounded-md px-2 text-xs font-mono",
            "bg-white/5 text-foreground border border-white/10",
            "transition-all duration-200",
            "hover:border-white/15 focus:border-gold/40 focus:outline-none focus:ring-1 focus:ring-gold/20",
            "cursor-pointer appearance-none"
          )}
        >
          {riskLevels.map((r) => (
            <option
              key={r.value}
              value={r.value}
              className="bg-surface text-foreground"
            >
              {r.label}
            </option>
          ))}
        </select>
      </div>

      {/* Police Station Input with Suggestions */}
      <div className="flex items-center gap-1.5 relative" ref={containerRef}>
        <Building2 className="h-3.5 w-3.5 text-muted" />
        <div className="relative">
          <input
            type="text"
            placeholder="Station..."
            value={inputValue}
            onFocus={() => setIsOpen(true)}
            onChange={(e) => {
              const val = e.target.value;
              setInputValue(val);
              setIsOpen(true);
              if (val === "") {
                setSelectedPoliceStation(null);
              }
            }}
            className={cn(
              "h-8 w-28 rounded-md px-2 text-xs font-mono",
              "bg-white/5 text-foreground placeholder:text-muted/50 border border-white/10",
              "transition-all duration-200",
              "hover:border-white/15 focus:border-gold/40 focus:outline-none focus:ring-1 focus:ring-gold/20"
            )}
          />
          {suggestions.length > 0 && (
            <div className="absolute top-full left-0 mt-1 w-44 max-h-48 overflow-y-auto z-50 rounded-md border border-white/10 bg-[#121617]/95 backdrop-blur-md shadow-2xl py-1">
              {suggestions.map((station) => (
                <button
                  key={station}
                  onClick={() => handleSelect(station)}
                  className={cn(
                    "w-full text-left px-3 py-1.5 text-[11px] font-mono transition-all duration-150",
                    station === selectedPoliceStation
                      ? "bg-gold/25 text-gold font-semibold"
                      : "text-foreground hover:bg-white/5 hover:text-gold"
                  )}
                >
                  {station}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Reset */}
      <button
        onClick={resetFilters}
        className={cn(
          "h-8 px-2 rounded-md text-xs text-muted",
          "hover:text-foreground hover:bg-white/5",
          "transition-all duration-200 active:scale-95"
        )}
        title="Reset filters"
      >
        <RotateCcw className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}
