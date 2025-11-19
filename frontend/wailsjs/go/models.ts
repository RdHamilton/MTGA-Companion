export namespace main {
	
	export class CardRatingWithTier {
	    name: string;
	    color: string;
	    rarity: string;
	    mtga_id?: number;
	    ever_drawn_win_rate: number;
	    opening_hand_win_rate: number;
	    ever_drawn_game_win_rate: number;
	    drawn_win_rate: number;
	    in_hand_win_rate: number;
	    ever_drawn_improvement_win_rate: number;
	    opening_hand_improvement_win_rate: number;
	    drawn_improvement_win_rate: number;
	    in_hand_improvement_win_rate: number;
	    avg_seen: number;
	    avg_pick: number;
	    pick_rate?: number;
	    "# ever_drawn": number;
	    "# opening_hand": number;
	    "# games": number;
	    "# drawn": number;
	    "# in_hand_drawn": number;
	    "# games_played"?: number;
	    "# decks"?: number;
	    tier: string;
	
	    static createFrom(source: any = {}) {
	        return new CardRatingWithTier(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.color = source["color"];
	        this.rarity = source["rarity"];
	        this.mtga_id = source["mtga_id"];
	        this.ever_drawn_win_rate = source["ever_drawn_win_rate"];
	        this.opening_hand_win_rate = source["opening_hand_win_rate"];
	        this.ever_drawn_game_win_rate = source["ever_drawn_game_win_rate"];
	        this.drawn_win_rate = source["drawn_win_rate"];
	        this.in_hand_win_rate = source["in_hand_win_rate"];
	        this.ever_drawn_improvement_win_rate = source["ever_drawn_improvement_win_rate"];
	        this.opening_hand_improvement_win_rate = source["opening_hand_improvement_win_rate"];
	        this.drawn_improvement_win_rate = source["drawn_improvement_win_rate"];
	        this.in_hand_improvement_win_rate = source["in_hand_improvement_win_rate"];
	        this.avg_seen = source["avg_seen"];
	        this.avg_pick = source["avg_pick"];
	        this.pick_rate = source["pick_rate"];
	        this["# ever_drawn"] = source["# ever_drawn"];
	        this["# opening_hand"] = source["# opening_hand"];
	        this["# games"] = source["# games"];
	        this["# drawn"] = source["# drawn"];
	        this["# in_hand_drawn"] = source["# in_hand_drawn"];
	        this["# games_played"] = source["# games_played"];
	        this["# decks"] = source["# decks"];
	        this.tier = source["tier"];
	    }
	}

}

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
	export class DraftEvent {
	    ID: string;
	    AccountID: number;
	    EventName: string;
	    SetCode: string;
	    StartTime: time.Time;
	    EndTime?: time.Time;
	    Wins: number;
	    Losses: number;
	    Status: string;
	    DeckID?: string;
	    EntryFee?: string;
	    Rewards?: string;
	    CreatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new DraftEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.AccountID = source["AccountID"];
	        this.EventName = source["EventName"];
	        this.SetCode = source["SetCode"];
	        this.StartTime = this.convertValues(source["StartTime"], time.Time);
	        this.EndTime = this.convertValues(source["EndTime"], time.Time);
	        this.Wins = source["Wins"];
	        this.Losses = source["Losses"];
	        this.Status = source["Status"];
	        this.DeckID = source["DeckID"];
	        this.EntryFee = source["EntryFee"];
	        this.Rewards = source["Rewards"];
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
	export class DraftPackSession {
	    ID: number;
	    SessionID: string;
	    PackNumber: number;
	    PickNumber: number;
	    CardIDs: string[];
	    Timestamp: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new DraftPackSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.SessionID = source["SessionID"];
	        this.PackNumber = source["PackNumber"];
	        this.PickNumber = source["PickNumber"];
	        this.CardIDs = source["CardIDs"];
	        this.Timestamp = this.convertValues(source["Timestamp"], time.Time);
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
	export class DraftPickSession {
	    ID: number;
	    SessionID: string;
	    PackNumber: number;
	    PickNumber: number;
	    CardID: string;
	    Timestamp: time.Time;
	    PickQualityGrade?: string;
	    PickQualityRank?: number;
	    PackBestGIHWR?: number;
	    PickedCardGIHWR?: number;
	    AlternativesJSON?: string;
	
	    static createFrom(source: any = {}) {
	        return new DraftPickSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.SessionID = source["SessionID"];
	        this.PackNumber = source["PackNumber"];
	        this.PickNumber = source["PickNumber"];
	        this.CardID = source["CardID"];
	        this.Timestamp = this.convertValues(source["Timestamp"], time.Time);
	        this.PickQualityGrade = source["PickQualityGrade"];
	        this.PickQualityRank = source["PickQualityRank"];
	        this.PackBestGIHWR = source["PackBestGIHWR"];
	        this.PickedCardGIHWR = source["PickedCardGIHWR"];
	        this.AlternativesJSON = source["AlternativesJSON"];
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
	export class DraftSession {
	    ID: string;
	    EventName: string;
	    SetCode: string;
	    DraftType: string;
	    StartTime: time.Time;
	    EndTime?: time.Time;
	    Status: string;
	    TotalPicks: number;
	    CreatedAt: time.Time;
	    UpdatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new DraftSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.EventName = source["EventName"];
	        this.SetCode = source["SetCode"];
	        this.DraftType = source["DraftType"];
	        this.StartTime = this.convertValues(source["StartTime"], time.Time);
	        this.EndTime = this.convertValues(source["EndTime"], time.Time);
	        this.Status = source["Status"];
	        this.TotalPicks = source["TotalPicks"];
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
	export class SetCard {
	    ID: number;
	    SetCode: string;
	    ArenaID: string;
	    ScryfallID: string;
	    Name: string;
	    ManaCost: string;
	    CMC: number;
	    Types: string[];
	    Colors: string[];
	    Rarity: string;
	    Text: string;
	    Power: string;
	    Toughness: string;
	    ImageURL: string;
	    ImageURLSmall: string;
	    ImageURLArt: string;
	    FetchedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new SetCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.SetCode = source["SetCode"];
	        this.ArenaID = source["ArenaID"];
	        this.ScryfallID = source["ScryfallID"];
	        this.Name = source["Name"];
	        this.ManaCost = source["ManaCost"];
	        this.CMC = source["CMC"];
	        this.Types = source["Types"];
	        this.Colors = source["Colors"];
	        this.Rarity = source["Rarity"];
	        this.Text = source["Text"];
	        this.Power = source["Power"];
	        this.Toughness = source["Toughness"];
	        this.ImageURL = source["ImageURL"];
	        this.ImageURLSmall = source["ImageURLSmall"];
	        this.ImageURLArt = source["ImageURLArt"];
	        this.FetchedAt = this.convertValues(source["FetchedAt"], time.Time);
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
	    DeckFormat?: string;
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
	        this.DeckFormat = source["DeckFormat"];
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

export namespace pickquality {
	
	export class Alternative {
	    card_id: string;
	    card_name: string;
	    gihwr: number;
	    rank: number;
	
	    static createFrom(source: any = {}) {
	        return new Alternative(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.card_id = source["card_id"];
	        this.card_name = source["card_name"];
	        this.gihwr = source["gihwr"];
	        this.rank = source["rank"];
	    }
	}
	export class PickQuality {
	    grade: string;
	    rank: number;
	    pack_best_gihwr: number;
	    picked_card_gihwr: number;
	    alternatives: Alternative[];
	
	    static createFrom(source: any = {}) {
	        return new PickQuality(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.grade = source["grade"];
	        this.rank = source["rank"];
	        this.pack_best_gihwr = source["pack_best_gihwr"];
	        this.picked_card_gihwr = source["picked_card_gihwr"];
	        this.alternatives = this.convertValues(source["alternatives"], Alternative);
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

export namespace seventeenlands {
	
	export class ColorRating {
	    color_name: string;
	    colors?: string[];
	    is_splash?: boolean;
	    splash_color?: string;
	    win_rate: number;
	    match_win_rate?: number;
	    game_win_rate?: number;
	    "# games": number;
	    "# matches"?: number;
	    "# wins"?: number;
	    "# losses"?: number;
	    "# decks"?: number;
	    avg_mainboard?: number;
	    avg_sideboard?: number;
	
	    static createFrom(source: any = {}) {
	        return new ColorRating(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.color_name = source["color_name"];
	        this.colors = source["colors"];
	        this.is_splash = source["is_splash"];
	        this.splash_color = source["splash_color"];
	        this.win_rate = source["win_rate"];
	        this.match_win_rate = source["match_win_rate"];
	        this.game_win_rate = source["game_win_rate"];
	        this["# games"] = source["# games"];
	        this["# matches"] = source["# matches"];
	        this["# wins"] = source["# wins"];
	        this["# losses"] = source["# losses"];
	        this["# decks"] = source["# decks"];
	        this.avg_mainboard = source["avg_mainboard"];
	        this.avg_sideboard = source["avg_sideboard"];
	    }
	}

}

export namespace storage {
	
	export class EventWinDistribution {
	    record: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new EventWinDistribution(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.record = source["record"];
	        this.count = source["count"];
	    }
	}
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

