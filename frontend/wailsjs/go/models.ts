export namespace models {
	
	export class Account {
	    ID: number;
	    Name: string;
	    ScreenName?: string;
	    ClientID?: string;
	    DailyWins: number;
	    WeeklyWins: number;
	    MasteryLevel: number;
	    MasteryPass: string;
	    MasteryMax: number;
	    IsDefault: boolean;
	    CreatedAt: time.Time;
	    UpdatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.ScreenName = source["ScreenName"];
	        this.ClientID = source["ClientID"];
	        this.DailyWins = source["DailyWins"];
	        this.WeeklyWins = source["WeeklyWins"];
	        this.MasteryLevel = source["MasteryLevel"];
	        this.MasteryPass = source["MasteryPass"];
	        this.MasteryMax = source["MasteryMax"];
	        this.IsDefault = source["IsDefault"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], time.Time);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Achievement {
	    ID: number;
	    AccountID: number;
	    GraphID: string;
	    NodeID: string;
	    Status: string;
	    CurrentProgress: number;
	    MaxProgress?: number;
	    CompletedAt?: time.Time;
	    FirstSeen: time.Time;
	    LastUpdated: time.Time;
	    CreatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Achievement(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.AccountID = source["AccountID"];
	        this.GraphID = source["GraphID"];
	        this.NodeID = source["NodeID"];
	        this.Status = source["Status"];
	        this.CurrentProgress = source["CurrentProgress"];
	        this.MaxProgress = source["MaxProgress"];
	        this.CompletedAt = this.convertValues(source["CompletedAt"], time.Time);
	        this.FirstSeen = this.convertValues(source["FirstSeen"], time.Time);
	        this.LastUpdated = this.convertValues(source["LastUpdated"], time.Time);
	        this.CreatedAt = this.convertValues(source["CreatedAt"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AchievementStats {
	    TotalAchievements: number;
	    CompletedAchievements: number;
	    InProgressCount: number;
	    CompletionRate: number;
	    RecentlyCompleted: number;
	    CloseToComplete: number;
	
	    static createFrom(source: any = {}) {
	        return new AchievementStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TotalAchievements = source["TotalAchievements"];
	        this.CompletedAchievements = source["CompletedAchievements"];
	        this.InProgressCount = source["InProgressCount"];
	        this.CompletionRate = source["CompletionRate"];
	        this.RecentlyCompleted = source["RecentlyCompleted"];
	        this.CloseToComplete = source["CloseToComplete"];
	    }
	}
	export class Match {
	    ID: string;
	    AccountID: number;
	    EventID: string;
	    EventName: string;
	    Timestamp: time.Time;
	    DurationSeconds?: number;
	    PlayerWins: number;
	    OpponentWins: number;
	    PlayerTeamID: number;
	    DeckID?: string;
	    RankBefore?: string;
	    RankAfter?: string;
	    Format: string;
	    Result: string;
	    ResultReason?: string;
	    OpponentName?: string;
	    OpponentID?: string;
	    CreatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Match(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.AccountID = source["AccountID"];
	        this.EventID = source["EventID"];
	        this.EventName = source["EventName"];
	        this.Timestamp = this.convertValues(source["Timestamp"], time.Time);
	        this.DurationSeconds = source["DurationSeconds"];
	        this.PlayerWins = source["PlayerWins"];
	        this.OpponentWins = source["OpponentWins"];
	        this.PlayerTeamID = source["PlayerTeamID"];
	        this.DeckID = source["DeckID"];
	        this.RankBefore = source["RankBefore"];
	        this.RankAfter = source["RankAfter"];
	        this.Format = source["Format"];
	        this.Result = source["Result"];
	        this.ResultReason = source["ResultReason"];
	        this.OpponentName = source["OpponentName"];
	        this.OpponentID = source["OpponentID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PerformanceMetrics {
	    AvgMatchDuration?: number;
	    AvgGameDuration?: number;
	    FastestMatch?: number;
	    SlowestMatch?: number;
	    FastestGame?: number;
	    SlowestGame?: number;
	
	    static createFrom(source: any = {}) {
	        return new PerformanceMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AvgMatchDuration = source["AvgMatchDuration"];
	        this.AvgGameDuration = source["AvgGameDuration"];
	        this.FastestMatch = source["FastestMatch"];
	        this.SlowestMatch = source["SlowestMatch"];
	        this.FastestGame = source["FastestGame"];
	        this.SlowestGame = source["SlowestGame"];
	    }
	}
	export class Quest {
	    id: number;
	    quest_id: string;
	    quest_type: string;
	    goal: number;
	    starting_progress: number;
	    ending_progress: number;
	    completed: boolean;
	    can_swap: boolean;
	    rewards: string;
	    assigned_at: time.Time;
	    completed_at?: time.Time;
	    rerolled: boolean;
	    created_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Quest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.quest_id = source["quest_id"];
	        this.quest_type = source["quest_type"];
	        this.goal = source["goal"];
	        this.starting_progress = source["starting_progress"];
	        this.ending_progress = source["ending_progress"];
	        this.completed = source["completed"];
	        this.can_swap = source["can_swap"];
	        this.rewards = source["rewards"];
	        this.assigned_at = this.convertValues(source["assigned_at"], time.Time);
	        this.completed_at = this.convertValues(source["completed_at"], time.Time);
	        this.rerolled = source["rerolled"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class QuestStats {
	    total_quests: number;
	    completed_quests: number;
	    active_quests: number;
	    completion_rate: number;
	    total_gold_earned: number;
	    average_completion_ms: number;
	    reroll_count: number;
	
	    static createFrom(source: any = {}) {
	        return new QuestStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_quests = source["total_quests"];
	        this.completed_quests = source["completed_quests"];
	        this.active_quests = source["active_quests"];
	        this.completion_rate = source["completion_rate"];
	        this.total_gold_earned = source["total_gold_earned"];
	        this.average_completion_ms = source["average_completion_ms"];
	        this.reroll_count = source["reroll_count"];
	    }
	}
	export class RankProgression {
	    CurrentRank: string;
	    NextRank: string;
	    CurrentStep: number;
	    StepsToNext: number;
	    IsAtFloor: boolean;
	    EstimatedMatches?: number;
	    WinRateUsed?: number;
	    Format: string;
	    LastUpdated: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new RankProgression(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CurrentRank = source["CurrentRank"];
	        this.NextRank = source["NextRank"];
	        this.CurrentStep = source["CurrentStep"];
	        this.StepsToNext = source["StepsToNext"];
	        this.IsAtFloor = source["IsAtFloor"];
	        this.EstimatedMatches = source["EstimatedMatches"];
	        this.WinRateUsed = source["WinRateUsed"];
	        this.Format = source["Format"];
	        this.LastUpdated = this.convertValues(source["LastUpdated"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Statistics {
	    TotalMatches: number;
	    MatchesWon: number;
	    MatchesLost: number;
	    TotalGames: number;
	    GamesWon: number;
	    GamesLost: number;
	    WinRate: number;
	    GameWinRate: number;
	
	    static createFrom(source: any = {}) {
	        return new Statistics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TotalMatches = source["TotalMatches"];
	        this.MatchesWon = source["MatchesWon"];
	        this.MatchesLost = source["MatchesLost"];
	        this.TotalGames = source["TotalGames"];
	        this.GamesWon = source["GamesWon"];
	        this.GamesLost = source["GamesLost"];
	        this.WinRate = source["WinRate"];
	        this.GameWinRate = source["GameWinRate"];
	    }
	}
	export class StatsFilter {
	    AccountID?: number;
	    StartDate?: time.Time;
	    EndDate?: time.Time;
	    Format?: string;
	    Formats: string[];
	    DeckID?: string;
	    EventName?: string;
	    EventNames: string[];
	    OpponentName?: string;
	    OpponentID?: string;
	    Result?: string;
	    RankClass?: string;
	    RankMinClass?: string;
	    RankMaxClass?: string;
	    ResultReason?: string;
	
	    static createFrom(source: any = {}) {
	        return new StatsFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AccountID = source["AccountID"];
	        this.StartDate = this.convertValues(source["StartDate"], time.Time);
	        this.EndDate = this.convertValues(source["EndDate"], time.Time);
	        this.Format = source["Format"];
	        this.Formats = source["Formats"];
	        this.DeckID = source["DeckID"];
	        this.EventName = source["EventName"];
	        this.EventNames = source["EventNames"];
	        this.OpponentName = source["OpponentName"];
	        this.OpponentID = source["OpponentID"];
	        this.Result = source["Result"];
	        this.RankClass = source["RankClass"];
	        this.RankMinClass = source["RankMinClass"];
	        this.RankMaxClass = source["RankMaxClass"];
	        this.ResultReason = source["ResultReason"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace storage {
	
	export class RankTimelineEntry {
	    timestamp: time.Time;
	    date: string;
	    rank: string;
	    rank_class?: string;
	    rank_level?: number;
	    rank_step?: number;
	    percentile?: number;
	    format: string;
	    season_ordinal: number;
	    is_change: boolean;
	    is_milestone: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RankTimelineEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], time.Time);
	        this.date = source["date"];
	        this.rank = source["rank"];
	        this.rank_class = source["rank_class"];
	        this.rank_level = source["rank_level"];
	        this.rank_step = source["rank_step"];
	        this.percentile = source["percentile"];
	        this.format = source["format"];
	        this.season_ordinal = source["season_ordinal"];
	        this.is_change = source["is_change"];
	        this.is_milestone = source["is_milestone"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RankTimeline {
	    format: string;
	    start_date: time.Time;
	    end_date: time.Time;
	    entries: RankTimelineEntry[];
	    total_changes: number;
	    milestones: number;
	    start_rank: string;
	    end_rank: string;
	    highest_rank: string;
	    lowest_rank: string;
	    seasons_covered: number[];
	
	    static createFrom(source: any = {}) {
	        return new RankTimeline(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.start_date = this.convertValues(source["start_date"], time.Time);
	        this.end_date = this.convertValues(source["end_date"], time.Time);
	        this.entries = this.convertValues(source["entries"], RankTimelineEntry);
	        this.total_changes = source["total_changes"];
	        this.milestones = source["milestones"];
	        this.start_rank = source["start_rank"];
	        this.end_rank = source["end_rank"];
	        this.highest_rank = source["highest_rank"];
	        this.lowest_rank = source["lowest_rank"];
	        this.seasons_covered = source["seasons_covered"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class TrendPeriod {
	    StartDate: time.Time;
	    EndDate: time.Time;
	    Label: string;
	
	    static createFrom(source: any = {}) {
	        return new TrendPeriod(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.StartDate = this.convertValues(source["StartDate"], time.Time);
	        this.EndDate = this.convertValues(source["EndDate"], time.Time);
	        this.Label = source["Label"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TrendData {
	    Period: TrendPeriod;
	    Stats?: models.Statistics;
	    WinRate: number;
	    GameWinRate: number;
	
	    static createFrom(source: any = {}) {
	        return new TrendData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Period = this.convertValues(source["Period"], TrendPeriod);
	        this.Stats = this.convertValues(source["Stats"], models.Statistics);
	        this.WinRate = source["WinRate"];
	        this.GameWinRate = source["GameWinRate"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TrendAnalysis {
	    Periods: TrendData[];
	    Overall?: models.Statistics;
	    Trend: string;
	    TrendValue: number;
	
	    static createFrom(source: any = {}) {
	        return new TrendAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Periods = this.convertValues(source["Periods"], TrendData);
	        this.Overall = this.convertValues(source["Overall"], models.Statistics);
	        this.Trend = source["Trend"];
	        this.TrendValue = source["TrendValue"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace time {
	
	export class Time {
	
	
	    static createFrom(source: any = {}) {
	        return new Time(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

