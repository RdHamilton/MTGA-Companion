// Match types
export interface Match {
  id: string;
  timestamp: string;
  result: 'win' | 'loss';
  format: string;
  event_name: string;
  opponent_name: string;
  wins: number;
  losses: number;
  rank_class: string;
  rank_level: number;
  deck_id: string;
}

// Statistics types
export interface Statistics {
  total_matches: number;
  matches_won: number;
  matches_lost: number;
  win_rate: number;
  total_games: number;
  games_won: number;
  games_lost: number;
  game_win_rate: number;
}

// Trend Analysis types
export interface TrendPeriod {
  start_date: string;
  end_date: string;
  label: string;
}

export interface TrendData {
  period: TrendPeriod;
  stats: Statistics;
  win_rate: number;
  game_win_rate: number;
}

export interface TrendAnalysis {
  periods: TrendData[];
  overall: Statistics | null;
  trend: string;
  trend_value: number;
}

// Filter types
export interface StatsFilter {
  start_date?: string | null;
  end_date?: string | null;
  format?: string | null;
  formats?: string[] | null;
  event_name?: string | null;
  opponent?: string | null;
  result?: string | null;
}
