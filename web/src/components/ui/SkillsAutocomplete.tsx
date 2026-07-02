import { X } from 'lucide-react';
import { useCallback, useEffect, useId, useRef, useState } from 'react';
import { getPopularSkills, searchSkills } from '@/lib/skills';

interface Props {
  value: string;
  onChange: (value: string) => void;
  label?: string;
  placeholder?: string;
}

export function SkillsAutocomplete({ value, onChange, label, placeholder }: Props) {
  const [inputValue, setInputValue] = useState('');
  const [suggestions, setSuggestions] = useState<{ id: string; name: string }[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const inputId = useId();

  const selectedSkills = value
    ? value
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
    : [];

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const fetchSuggestions = useCallback(async (q: string) => {
    if (!q.trim()) {
      try {
        const popular = await getPopularSkills();
        setSuggestions(popular);
      } catch {
        setSuggestions([]);
      }
      return;
    }
    setLoading(true);
    try {
      const results = await searchSkills(q);
      setSuggestions(results);
    } catch {
      setSuggestions([]);
    } finally {
      setLoading(false);
    }
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setInputValue(val);
    setIsOpen(true);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => fetchSuggestions(val), 300);
  };

  const addSkill = (skill: string) => {
    const trimmed = skill.trim();
    if (!trimmed || selectedSkills.includes(trimmed)) return;
    const newVal = [...selectedSkills, trimmed].join(', ');
    onChange(newVal);
    setInputValue('');
    setIsOpen(false);
  };

  const removeSkill = (skill: string) => {
    const newVal = selectedSkills.filter((s) => s !== skill).join(', ');
    onChange(newVal);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && inputValue.trim()) {
      e.preventDefault();
      addSkill(inputValue);
    }
    if (e.key === 'Backspace' && !inputValue && selectedSkills.length > 0) {
      removeSkill(selectedSkills[selectedSkills.length - 1]);
    }
  };

  return (
    <div className="form-control relative" ref={wrapperRef}>
      {label && (
        <label className="label" htmlFor={inputId}>
          <span className="label-text">{label}</span>
        </label>
      )}
      <div className="flex flex-wrap gap-1 mb-1">
        {selectedSkills.map((skill) => (
          <span key={skill} className="badge badge-primary gap-1">
            {skill}
            <button
              type="button"
              onClick={() => removeSkill(skill)}
              className="btn btn-xs btn-ghost btn-circle p-0"
              aria-label={`Remove ${skill}`}
            >
              <X size={12} />
            </button>
          </span>
        ))}
      </div>
      <input
        type="text"
        id={inputId}
        className="input input-bordered w-full"
        value={inputValue}
        onChange={handleInputChange}
        onFocus={() => {
          setIsOpen(true);
          fetchSuggestions(inputValue);
        }}
        onKeyDown={handleKeyDown}
        placeholder={placeholder ?? 'Type a skill and press Enter'}
      />
      {isOpen && suggestions.length > 0 && (
        <ul className="menu bg-base-100 rounded-box shadow-sm z-10 max-h-40 overflow-y-auto w-full absolute mt-1">
          {suggestions.map((s) => (
            <li key={s.id}>
              <button
                type="button"
                className="text-sm w-full text-left px-3 py-1.5 hover:bg-base-200"
                onClick={() => addSkill(s.name)}
              >
                {s.name}
              </button>
            </li>
          ))}
        </ul>
      )}
      {loading && <span className="loading loading-spinner loading-xs mt-1" />}
    </div>
  );
}
