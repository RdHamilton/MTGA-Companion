export namespace grading {
	
	export class DraftGrade {
	    overall_grade: string;
	    overall_score: number;
	    pick_quality_score: number;
	    color_discipline_score: number;
	    deck_composition_score: number;
	    strategic_score: number;
	    best_picks: string[];
	    worst_picks: string[];
	    suggestions: string[];
	
	    static createFrom(source: any = {}) {
	        return new DraftGrade(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.overall_grade = source["overall_grade"];
	        this.overall_score = source["overall_score"];
	        this.pick_quality_score = source["pick_quality_score"];
	        this.color_discipline_score = source["color_discipline_score"];
	        this.deck_composition_score = source["deck_composition_score"];
	        this.strategic_score = source["strategic_score"];
	        this.best_picks = source["best_picks"];
	        this.worst_picks = source["worst_picks"];
	        this.suggestions = source["suggestions"];
	    }
	}

}

export namespace gui {
	
	export class AppSettings {
	    autoRefresh: boolean;
	    refreshInterval: number;
	    showNotifications: boolean;
	    theme: string;
	    daemonPort: number;
	    daemonMode: string;
	    mlEnabled: boolean;
	    llmEnabled: boolean;
	    ollamaEndpoint: string;
	    ollamaModel: string;
	    metaGoldfishEnabled: boolean;
	    metaTop8Enabled: boolean;
	    metaWeight: number;
	    personalWeight: number;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.autoRefresh = source["autoRefresh"];
	        this.refreshInterval = source["refreshInterval"];
	        this.showNotifications = source["showNotifications"];
	        this.theme = source["theme"];
	        this.daemonPort = source["daemonPort"];
	        this.daemonMode = source["daemonMode"];
	        this.mlEnabled = source["mlEnabled"];
	        this.llmEnabled = source["llmEnabled"];
	        this.ollamaEndpoint = source["ollamaEndpoint"];
	        this.ollamaModel = source["ollamaModel"];
	        this.metaGoldfishEnabled = source["metaGoldfishEnabled"];
	        this.metaTop8Enabled = source["metaTop8Enabled"];
	        this.metaWeight = source["metaWeight"];
	        this.personalWeight = source["personalWeight"];
	    }
	}
	export class DeckArchetypeAnalysis {
	    colorCounts: Record<string, number>;
	    colorlessCount: number;
	    goldCount: number;
	    creatureCount: number;
	    instantCount: number;
	    sorceryCount: number;
	    artifactCount: number;
	    enchantmentCount: number;
	    planeswalkerCount: number;
	    landCount: number;
	    manaCurve: Record<number, number>;
	    avgCMC: number;
	    rareCounts: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new DeckArchetypeAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.colorCounts = source["colorCounts"];
	        this.colorlessCount = source["colorlessCount"];
	        this.goldCount = source["goldCount"];
	        this.creatureCount = source["creatureCount"];
	        this.instantCount = source["instantCount"];
	        this.sorceryCount = source["sorceryCount"];
	        this.artifactCount = source["artifactCount"];
	        this.enchantmentCount = source["enchantmentCount"];
	        this.planeswalkerCount = source["planeswalkerCount"];
	        this.landCount = source["landCount"];
	        this.manaCurve = source["manaCurve"];
	        this.avgCMC = source["avgCMC"];
	        this.rareCounts = source["rareCounts"];
	    }
	}
	export class ArchetypeIndicatorInfo {
	    cardID: number;
	    cardName: string;
	    weight: number;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new ArchetypeIndicatorInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cardID = source["cardID"];
	        this.cardName = source["cardName"];
	        this.weight = source["weight"];
	        this.reason = source["reason"];
	    }
	}
	export class ColorPairInfo {
	    colors: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ColorPairInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.colors = source["colors"];
	        this.name = source["name"];
	    }
	}
	export class ArchetypeClassificationResult {
	    primaryArchetype: string;
	    secondaryArchetype?: string;
	    confidence: number;
	    confidencePercent: number;
	    colorIdentity: string;
	    dominantColors: string[];
	    colorPair?: ColorPairInfo;
	    signatureCards: number[];
	    indicators: ArchetypeIndicatorInfo[];
	    totalCards: number;
	    analysis?: DeckArchetypeAnalysis;
	
	    static createFrom(source: any = {}) {
	        return new ArchetypeClassificationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.primaryArchetype = source["primaryArchetype"];
	        this.secondaryArchetype = source["secondaryArchetype"];
	        this.confidence = source["confidence"];
	        this.confidencePercent = source["confidencePercent"];
	        this.colorIdentity = source["colorIdentity"];
	        this.dominantColors = source["dominantColors"];
	        this.colorPair = this.convertValues(source["colorPair"], ColorPairInfo);
	        this.signatureCards = source["signatureCards"];
	        this.indicators = this.convertValues(source["indicators"], ArchetypeIndicatorInfo);
	        this.totalCards = source["totalCards"];
	        this.analysis = this.convertValues(source["analysis"], DeckArchetypeAnalysis);
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
	
	export class ArchetypeInfo {
	    name: string;
	    colors: string[];
	    metaShare: number;
	    tournamentTop8s: number;
	    tournamentWins: number;
	    tier: number;
	    confidenceScore: number;
	    trendDirection: string;
	
	    static createFrom(source: any = {}) {
	        return new ArchetypeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.colors = source["colors"];
	        this.metaShare = source["metaShare"];
	        this.tournamentTop8s = source["tournamentTop8s"];
	        this.tournamentWins = source["tournamentWins"];
	        this.tier = source["tier"];
	        this.confidenceScore = source["confidenceScore"];
	        this.trendDirection = source["trendDirection"];
	    }
	}
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
	    colors: string[];
	
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
	        this.colors = source["colors"];
	    }
	}
	export class ScoreFactors {
	    colorFit: number;
	    manaCurve: number;
	    synergy: number;
	    quality: number;
	    playable: number;
	
	    static createFrom(source: any = {}) {
	        return new ScoreFactors(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.colorFit = source["colorFit"];
	        this.manaCurve = source["manaCurve"];
	        this.synergy = source["synergy"];
	        this.quality = source["quality"];
	        this.playable = source["playable"];
	    }
	}
	export class CardRecommendation {
	    cardID: number;
	    name: string;
	    typeLine: string;
	    manaCost?: string;
	    imageURI?: string;
	    score: number;
	    reasoning: string;
	    source: string;
	    confidence: number;
	    factors?: ScoreFactors;
	
	    static createFrom(source: any = {}) {
	        return new CardRecommendation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cardID = source["cardID"];
	        this.name = source["name"];
	        this.typeLine = source["typeLine"];
	        this.manaCost = source["manaCost"];
	        this.imageURI = source["imageURI"];
	        this.score = source["score"];
	        this.reasoning = source["reasoning"];
	        this.source = source["source"];
	        this.confidence = source["confidence"];
	        this.factors = this.convertValues(source["factors"], ScoreFactors);
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
	export class CardWithOwned {
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
	    ownedQuantity: number;
	
	    static createFrom(source: any = {}) {
	        return new CardWithOwned(source);
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
	        this.ownedQuantity = source["ownedQuantity"];
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
	export class CollectionCard {
	    cardId: number;
	    arenaId: number;
	    quantity: number;
	    name: string;
	    setCode: string;
	    setName: string;
	    rarity: string;
	    manaCost: string;
	    cmc: number;
	    typeLine: string;
	    colors: string[];
	    colorIdentity: string[];
	    imageUri: string;
	    power?: string;
	    toughness?: string;
	
	    static createFrom(source: any = {}) {
	        return new CollectionCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cardId = source["cardId"];
	        this.arenaId = source["arenaId"];
	        this.quantity = source["quantity"];
	        this.name = source["name"];
	        this.setCode = source["setCode"];
	        this.setName = source["setName"];
	        this.rarity = source["rarity"];
	        this.manaCost = source["manaCost"];
	        this.cmc = source["cmc"];
	        this.typeLine = source["typeLine"];
	        this.colors = source["colors"];
	        this.colorIdentity = source["colorIdentity"];
	        this.imageUri = source["imageUri"];
	        this.power = source["power"];
	        this.toughness = source["toughness"];
	    }
	}
	export class CollectionChangeEntry {
	    cardId: number;
	    cardName?: string;
	    setCode?: string;
	    rarity?: string;
	    quantityDelta: number;
	    quantityAfter: number;
	    timestamp: number;
	    source?: string;
	
	    static createFrom(source: any = {}) {
	        return new CollectionChangeEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cardId = source["cardId"];
	        this.cardName = source["cardName"];
	        this.setCode = source["setCode"];
	        this.rarity = source["rarity"];
	        this.quantityDelta = source["quantityDelta"];
	        this.quantityAfter = source["quantityAfter"];
	        this.timestamp = source["timestamp"];
	        this.source = source["source"];
	    }
	}
	export class CollectionFilter {
	    searchTerm?: string;
	    setCode?: string;
	    rarity?: string;
	    colors?: string[];
	    cardType?: string;
	    ownedOnly: boolean;
	    sortBy?: string;
	    sortDesc: boolean;
	    limit?: number;
	    offset?: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.searchTerm = source["searchTerm"];
	        this.setCode = source["setCode"];
	        this.rarity = source["rarity"];
	        this.colors = source["colors"];
	        this.cardType = source["cardType"];
	        this.ownedOnly = source["ownedOnly"];
	        this.sortBy = source["sortBy"];
	        this.sortDesc = source["sortDesc"];
	        this.limit = source["limit"];
	        this.offset = source["offset"];
	    }
	}
	export class CollectionResponse {
	    cards: CollectionCard[];
	    totalCount: number;
	    filterCount: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cards = this.convertValues(source["cards"], CollectionCard);
	        this.totalCount = source["totalCount"];
	        this.filterCount = source["filterCount"];
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
	export class CollectionStats {
	    totalUniqueCards: number;
	    totalCards: number;
	    commonCount: number;
	    uncommonCount: number;
	    rareCount: number;
	    mythicCount: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalUniqueCards = source["totalUniqueCards"];
	        this.totalCards = source["totalCards"];
	        this.commonCount = source["commonCount"];
	        this.uncommonCount = source["uncommonCount"];
	        this.rareCount = source["rareCount"];
	        this.mythicCount = source["mythicCount"];
	    }
	}
	export class CollectionUpdatedEvent {
	    newCards: number;
	    cardsAdded: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionUpdatedEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.newCards = source["newCards"];
	        this.cardsAdded = source["cardsAdded"];
	    }
	}
	
	export class ColorStats {
	    white: number;
	    blue: number;
	    black: number;
	    red: number;
	    green: number;
	    colorless: number;
	    multicolor: number;
	
	    static createFrom(source: any = {}) {
	        return new ColorStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.white = source["white"];
	        this.blue = source["blue"];
	        this.black = source["black"];
	        this.red = source["red"];
	        this.green = source["green"];
	        this.colorless = source["colorless"];
	        this.multicolor = source["multicolor"];
	    }
	}
	export class ConnectionStatus {
	    status: string;
	    connected: boolean;
	    mode: string;
	    url: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.connected = source["connected"];
	        this.mode = source["mode"];
	        this.url = source["url"];
	        this.port = source["port"];
	    }
	}
	export class CreatureStats {
	    total: number;
	    averagePower: number;
	    averageToughness: number;
	    totalPower: number;
	    totalToughness: number;
	
	    static createFrom(source: any = {}) {
	        return new CreatureStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.averagePower = source["averagePower"];
	        this.averageToughness = source["averageToughness"];
	        this.totalPower = source["totalPower"];
	        this.totalToughness = source["totalToughness"];
	    }
	}
	export class PackCardWithRating {
	    arena_id: string;
	    name: string;
	    image_url: string;
	    rarity: string;
	    colors: string[];
	    mana_cost: string;
	    cmc: number;
	    type_line: string;
	    gihwr: number;
	    alsa: number;
	    tier: string;
	    is_recommended: boolean;
	    score: number;
	    reasoning: string;
	
	    static createFrom(source: any = {}) {
	        return new PackCardWithRating(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.arena_id = source["arena_id"];
	        this.name = source["name"];
	        this.image_url = source["image_url"];
	        this.rarity = source["rarity"];
	        this.colors = source["colors"];
	        this.mana_cost = source["mana_cost"];
	        this.cmc = source["cmc"];
	        this.type_line = source["type_line"];
	        this.gihwr = source["gihwr"];
	        this.alsa = source["alsa"];
	        this.tier = source["tier"];
	        this.is_recommended = source["is_recommended"];
	        this.score = source["score"];
	        this.reasoning = source["reasoning"];
	    }
	}
	export class CurrentPackResponse {
	    session_id: string;
	    pack_number: number;
	    pick_number: number;
	    pack_label: string;
	    cards: PackCardWithRating[];
	    recommended_card?: PackCardWithRating;
	    pool_colors: string[];
	    pool_size: number;
	
	    static createFrom(source: any = {}) {
	        return new CurrentPackResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session_id = source["session_id"];
	        this.pack_number = source["pack_number"];
	        this.pick_number = source["pick_number"];
	        this.pack_label = source["pack_label"];
	        this.cards = this.convertValues(source["cards"], PackCardWithRating);
	        this.recommended_card = this.convertValues(source["recommended_card"], PackCardWithRating);
	        this.pool_colors = source["pool_colors"];
	        this.pool_size = source["pool_size"];
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
	export class DaemonErrorEvent {
	    error: string;
	    code: string;
	    details: string;
	
	    static createFrom(source: any = {}) {
	        return new DaemonErrorEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.error = source["error"];
	        this.code = source["code"];
	        this.details = source["details"];
	    }
	}
	export class PeriodMetrics {
	    total: number;
	    acceptanceRate: number;
	    acceptancePercent: number;
	
	    static createFrom(source: any = {}) {
	        return new PeriodMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.acceptanceRate = source["acceptanceRate"];
	        this.acceptancePercent = source["acceptancePercent"];
	    }
	}
	export class TypeMetrics {
	    total: number;
	    acceptanceRate: number;
	    acceptancePercent: number;
	    winRateOnAccepted?: number;
	
	    static createFrom(source: any = {}) {
	        return new TypeMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.acceptanceRate = source["acceptanceRate"];
	        this.acceptancePercent = source["acceptancePercent"];
	        this.winRateOnAccepted = source["winRateOnAccepted"];
	    }
	}
	export class DashboardMetricsResponse {
	    totalRecommendations: number;
	    acceptanceRate: number;
	    acceptancePercent: number;
	    rejectionRate: number;
	    rejectionPercent: number;
	    winRateOnAccepted?: number;
	    winRateOnRejected?: number;
	    winRateDifference?: number;
	    byType: Record<string, TypeMetrics>;
	    last7Days?: PeriodMetrics;
	    last30Days?: PeriodMetrics;
	
	    static createFrom(source: any = {}) {
	        return new DashboardMetricsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalRecommendations = source["totalRecommendations"];
	        this.acceptanceRate = source["acceptanceRate"];
	        this.acceptancePercent = source["acceptancePercent"];
	        this.rejectionRate = source["rejectionRate"];
	        this.rejectionPercent = source["rejectionPercent"];
	        this.winRateOnAccepted = source["winRateOnAccepted"];
	        this.winRateOnRejected = source["winRateOnRejected"];
	        this.winRateDifference = source["winRateDifference"];
	        this.byType = this.convertValues(source["byType"], TypeMetrics, true);
	        this.last7Days = this.convertValues(source["last7Days"], PeriodMetrics);
	        this.last30Days = this.convertValues(source["last30Days"], PeriodMetrics);
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
	
	export class DeckLibraryFilter {
	    format?: string;
	    source?: string;
	    tags?: string[];
	    sortBy?: string;
	    sortDesc?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeckLibraryFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.source = source["source"];
	        this.tags = source["tags"];
	        this.sortBy = source["sortBy"];
	        this.sortDesc = source["sortDesc"];
	    }
	}
	export class DeckListItem {
	    id: string;
	    name: string;
	    format: string;
	    source: string;
	    colorIdentity?: string;
	    cardCount: number;
	    matchesPlayed: number;
	    matchWinRate: number;
	    modifiedAt: time.Time;
	    lastPlayed?: time.Time;
	    tags?: string[];
	    currentStreak: number;
	    averageDuration?: number;
	
	    static createFrom(source: any = {}) {
	        return new DeckListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.format = source["format"];
	        this.source = source["source"];
	        this.colorIdentity = source["colorIdentity"];
	        this.cardCount = source["cardCount"];
	        this.matchesPlayed = source["matchesPlayed"];
	        this.matchWinRate = source["matchWinRate"];
	        this.modifiedAt = this.convertValues(source["modifiedAt"], time.Time);
	        this.lastPlayed = this.convertValues(source["lastPlayed"], time.Time);
	        this.tags = source["tags"];
	        this.currentStreak = source["currentStreak"];
	        this.averageDuration = source["averageDuration"];
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
	export class LegalityStatus {
	    legal: boolean;
	    reasons?: string[];
	
	    static createFrom(source: any = {}) {
	        return new LegalityStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.legal = source["legal"];
	        this.reasons = source["reasons"];
	    }
	}
	export class FormatLegality {
	    standard: LegalityStatus;
	    historic: LegalityStatus;
	    explorer: LegalityStatus;
	    alchemy: LegalityStatus;
	    brawl: LegalityStatus;
	    commander: LegalityStatus;
	
	    static createFrom(source: any = {}) {
	        return new FormatLegality(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.standard = this.convertValues(source["standard"], LegalityStatus);
	        this.historic = this.convertValues(source["historic"], LegalityStatus);
	        this.explorer = this.convertValues(source["explorer"], LegalityStatus);
	        this.alchemy = this.convertValues(source["alchemy"], LegalityStatus);
	        this.brawl = this.convertValues(source["brawl"], LegalityStatus);
	        this.commander = this.convertValues(source["commander"], LegalityStatus);
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
	export class LandStats {
	    total: number;
	    basic: number;
	    nonBasic: number;
	    ratio: number;
	    recommended: number;
	    status: string;
	    statusMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new LandStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.basic = source["basic"];
	        this.nonBasic = source["nonBasic"];
	        this.ratio = source["ratio"];
	        this.recommended = source["recommended"];
	        this.status = source["status"];
	        this.statusMessage = source["statusMessage"];
	    }
	}
	export class TypeStats {
	    creatures: number;
	    instants: number;
	    sorceries: number;
	    enchantments: number;
	    artifacts: number;
	    planeswalkers: number;
	    lands: number;
	    other: number;
	
	    static createFrom(source: any = {}) {
	        return new TypeStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.creatures = source["creatures"];
	        this.instants = source["instants"];
	        this.sorceries = source["sorceries"];
	        this.enchantments = source["enchantments"];
	        this.artifacts = source["artifacts"];
	        this.planeswalkers = source["planeswalkers"];
	        this.lands = source["lands"];
	        this.other = source["other"];
	    }
	}
	export class DeckStatistics {
	    totalCards: number;
	    totalMainboard: number;
	    totalSideboard: number;
	    averageCMC: number;
	    manaCurve: Record<number, number>;
	    maxCMC: number;
	    colors: ColorStats;
	    types: TypeStats;
	    lands: LandStats;
	    creatures: CreatureStats;
	    legality: FormatLegality;
	
	    static createFrom(source: any = {}) {
	        return new DeckStatistics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalCards = source["totalCards"];
	        this.totalMainboard = source["totalMainboard"];
	        this.totalSideboard = source["totalSideboard"];
	        this.averageCMC = source["averageCMC"];
	        this.manaCurve = source["manaCurve"];
	        this.maxCMC = source["maxCMC"];
	        this.colors = this.convertValues(source["colors"], ColorStats);
	        this.types = this.convertValues(source["types"], TypeStats);
	        this.lands = this.convertValues(source["lands"], LandStats);
	        this.creatures = this.convertValues(source["creatures"], CreatureStats);
	        this.legality = this.convertValues(source["legality"], FormatLegality);
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
	export class DeckUpdatedEvent {
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new DeckUpdatedEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	    }
	}
	export class DeckWithCards {
	    deck?: models.Deck;
	    cards: models.DeckCard[];
	    tags?: models.DeckTag[];
	
	    static createFrom(source: any = {}) {
	        return new DeckWithCards(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deck = this.convertValues(source["deck"], models.Deck);
	        this.cards = this.convertValues(source["cards"], models.DeckCard);
	        this.tags = this.convertValues(source["tags"], models.DeckTag);
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
	export class DraftUpdatedEvent {
	    count: number;
	    picks: number;
	
	    static createFrom(source: any = {}) {
	        return new DraftUpdatedEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.picks = source["picks"];
	    }
	}
	export class ExplainRecommendationRequest {
	    deckID: string;
	    cardID: number;
	
	    static createFrom(source: any = {}) {
	        return new ExplainRecommendationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deckID = source["deckID"];
	        this.cardID = source["cardID"];
	    }
	}
	export class ExplainRecommendationResponse {
	    explanation: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ExplainRecommendationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.explanation = source["explanation"];
	        this.error = source["error"];
	    }
	}
	export class ExportDeckRequest {
	    deckID: string;
	    format: string;
	    includeHeaders: boolean;
	    includeStats: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ExportDeckRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deckID = source["deckID"];
	        this.format = source["format"];
	        this.includeHeaders = source["includeHeaders"];
	        this.includeStats = source["includeStats"];
	    }
	}
	export class ExportDeckResponse {
	    content: string;
	    filename: string;
	    format: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportDeckResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.filename = source["filename"];
	        this.format = source["format"];
	        this.error = source["error"];
	    }
	}
	
	export class GetRecommendationsRequest {
	    deckID: string;
	    maxResults?: number;
	    minScore?: number;
	    colors?: string[];
	    cardTypes?: string[];
	    cmcMin?: number;
	    cmcMax?: number;
	    includeLands: boolean;
	    onlyDraftPool?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GetRecommendationsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deckID = source["deckID"];
	        this.maxResults = source["maxResults"];
	        this.minScore = source["minScore"];
	        this.colors = source["colors"];
	        this.cardTypes = source["cardTypes"];
	        this.cmcMin = source["cmcMin"];
	        this.cmcMax = source["cmcMax"];
	        this.includeLands = source["includeLands"];
	        this.onlyDraftPool = source["onlyDraftPool"];
	    }
	}
	export class GetRecommendationsResponse {
	    recommendations: CardRecommendation[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new GetRecommendationsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.recommendations = this.convertValues(source["recommendations"], CardRecommendation);
	        this.error = source["error"];
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
	export class ImportDeckRequest {
	    name: string;
	    format: string;
	    importText: string;
	    source: string;
	    draftEventID?: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportDeckRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.format = source["format"];
	        this.importText = source["importText"];
	        this.source = source["source"];
	        this.draftEventID = source["draftEventID"];
	    }
	}
	export class ImportDeckResponse {
	    success: boolean;
	    deckID?: string;
	    errors?: string[];
	    warnings?: string[];
	    cardsImported: number;
	    cardsSkipped: number;
	
	    static createFrom(source: any = {}) {
	        return new ImportDeckResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.deckID = source["deckID"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	        this.cardsImported = source["cardsImported"];
	        this.cardsSkipped = source["cardsSkipped"];
	    }
	}
	export class ImportLogFileResult {
	    fileName: string;
	    entriesRead: number;
	    matchesStored: number;
	    gamesStored: number;
	    decksStored: number;
	    ranksStored: number;
	    questsStored: number;
	    draftsStored: number;
	    picksStored: number;
	
	    static createFrom(source: any = {}) {
	        return new ImportLogFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileName = source["fileName"];
	        this.entriesRead = source["entriesRead"];
	        this.matchesStored = source["matchesStored"];
	        this.gamesStored = source["gamesStored"];
	        this.decksStored = source["decksStored"];
	        this.ranksStored = source["ranksStored"];
	        this.questsStored = source["questsStored"];
	        this.draftsStored = source["draftsStored"];
	        this.picksStored = source["picksStored"];
	    }
	}
	
	
	export class LogReplayProgress {
	    totalFiles: number;
	    processedFiles: number;
	    currentFile: string;
	    totalEntries: number;
	    processedEntries: number;
	    percentComplete: number;
	    matchesImported: number;
	    decksImported: number;
	    questsImported: number;
	    draftsImported: number;
	    duration: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new LogReplayProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalFiles = source["totalFiles"];
	        this.processedFiles = source["processedFiles"];
	        this.currentFile = source["currentFile"];
	        this.totalEntries = source["totalEntries"];
	        this.processedEntries = source["processedEntries"];
	        this.percentComplete = source["percentComplete"];
	        this.matchesImported = source["matchesImported"];
	        this.decksImported = source["decksImported"];
	        this.questsImported = source["questsImported"];
	        this.draftsImported = source["draftsImported"];
	        this.duration = source["duration"];
	        this.error = source["error"];
	    }
	}
	export class RecommendationContextData {
	    deckID?: string;
	    draftEventID?: string;
	    format?: string;
	    setCode?: string;
	    deckCardCount: number;
	    deckColorIdentity?: string;
	    packNumber?: number;
	    pickNumber?: number;
	    availableCards?: number[];
	    currentArchetype?: string;
	    recommendedCards?: number[];
	
	    static createFrom(source: any = {}) {
	        return new RecommendationContextData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deckID = source["deckID"];
	        this.draftEventID = source["draftEventID"];
	        this.format = source["format"];
	        this.setCode = source["setCode"];
	        this.deckCardCount = source["deckCardCount"];
	        this.deckColorIdentity = source["deckColorIdentity"];
	        this.packNumber = source["packNumber"];
	        this.pickNumber = source["pickNumber"];
	        this.availableCards = source["availableCards"];
	        this.currentArchetype = source["currentArchetype"];
	        this.recommendedCards = source["recommendedCards"];
	    }
	}
	export class MLTrainingEntry {
	    recommendationType: string;
	    recommendedCardID?: number;
	    recommendedArchetype?: string;
	    context?: RecommendationContextData;
	    action: string;
	    alternateChoiceID?: number;
	    outcomeResult?: string;
	    recommendationScore?: number;
	    recommendationRank?: number;
	    recommendedAt: string;
	    respondedAt?: string;
	
	    static createFrom(source: any = {}) {
	        return new MLTrainingEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.recommendationType = source["recommendationType"];
	        this.recommendedCardID = source["recommendedCardID"];
	        this.recommendedArchetype = source["recommendedArchetype"];
	        this.context = this.convertValues(source["context"], RecommendationContextData);
	        this.action = source["action"];
	        this.alternateChoiceID = source["alternateChoiceID"];
	        this.outcomeResult = source["outcomeResult"];
	        this.recommendationScore = source["recommendationScore"];
	        this.recommendationRank = source["recommendationRank"];
	        this.recommendedAt = source["recommendedAt"];
	        this.respondedAt = source["respondedAt"];
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
	export class MLTrainingDataExport {
	    data: MLTrainingEntry[];
	    totalCount: number;
	    exportedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new MLTrainingDataExport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], MLTrainingEntry);
	        this.totalCount = source["totalCount"];
	        this.exportedAt = source["exportedAt"];
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
	
	export class TournamentInfo {
	    name: string;
	    date: time.Time;
	    players: number;
	    format: string;
	    topDecks: string[];
	    sourceUrl?: string;
	
	    static createFrom(source: any = {}) {
	        return new TournamentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.date = this.convertValues(source["date"], time.Time);
	        this.players = source["players"];
	        this.format = source["format"];
	        this.topDecks = source["topDecks"];
	        this.sourceUrl = source["sourceUrl"];
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
	export class MetaDashboardResponse {
	    format: string;
	    archetypes: ArchetypeInfo[];
	    tournaments?: TournamentInfo[];
	    totalArchetypes: number;
	    lastUpdated: time.Time;
	    sources: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new MetaDashboardResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.archetypes = this.convertValues(source["archetypes"], ArchetypeInfo);
	        this.tournaments = this.convertValues(source["tournaments"], TournamentInfo);
	        this.totalArchetypes = source["totalArchetypes"];
	        this.lastUpdated = this.convertValues(source["lastUpdated"], time.Time);
	        this.sources = source["sources"];
	        this.error = source["error"];
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
	export class MissingCard {
	    cardId: number;
	    arenaId: number;
	    name: string;
	    setCode: string;
	    setName: string;
	    rarity: string;
	    manaCost: string;
	    cmc: number;
	    typeLine: string;
	    colors: string[];
	    imageUri: string;
	    quantityNeeded: number;
	    quantityOwned: number;
	
	    static createFrom(source: any = {}) {
	        return new MissingCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cardId = source["cardId"];
	        this.arenaId = source["arenaId"];
	        this.name = source["name"];
	        this.setCode = source["setCode"];
	        this.setName = source["setName"];
	        this.rarity = source["rarity"];
	        this.manaCost = source["manaCost"];
	        this.cmc = source["cmc"];
	        this.typeLine = source["typeLine"];
	        this.colors = source["colors"];
	        this.imageUri = source["imageUri"];
	        this.quantityNeeded = source["quantityNeeded"];
	        this.quantityOwned = source["quantityOwned"];
	    }
	}
	export class WildcardCost {
	    common: number;
	    uncommon: number;
	    rare: number;
	    mythic: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new WildcardCost(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.common = source["common"];
	        this.uncommon = source["uncommon"];
	        this.rare = source["rare"];
	        this.mythic = source["mythic"];
	        this.total = source["total"];
	    }
	}
	export class MissingCardsForDeckResponse {
	    deckId: string;
	    deckName: string;
	    totalMissing: number;
	    uniqueMissing: number;
	    missingCards: MissingCard[];
	    wildcardsNeeded?: WildcardCost;
	    isComplete: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MissingCardsForDeckResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deckId = source["deckId"];
	        this.deckName = source["deckName"];
	        this.totalMissing = source["totalMissing"];
	        this.uniqueMissing = source["uniqueMissing"];
	        this.missingCards = this.convertValues(source["missingCards"], MissingCard);
	        this.wildcardsNeeded = this.convertValues(source["wildcardsNeeded"], WildcardCost);
	        this.isComplete = source["isComplete"];
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
	export class MissingCardsForSetResponse {
	    setCode: string;
	    setName: string;
	    totalMissing: number;
	    uniqueMissing: number;
	    missingCards: MissingCard[];
	    wildcardsNeeded?: WildcardCost;
	    completionPct: number;
	
	    static createFrom(source: any = {}) {
	        return new MissingCardsForSetResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.setCode = source["setCode"];
	        this.setName = source["setName"];
	        this.totalMissing = source["totalMissing"];
	        this.uniqueMissing = source["uniqueMissing"];
	        this.missingCards = this.convertValues(source["missingCards"], MissingCard);
	        this.wildcardsNeeded = this.convertValues(source["wildcardsNeeded"], WildcardCost);
	        this.completionPct = source["completionPct"];
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
	export class OllamaModel {
	    name: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new OllamaModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	    }
	}
	export class OllamaStatus {
	    available: boolean;
	    version?: string;
	    modelReady: boolean;
	    modelName: string;
	    modelsLoaded?: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new OllamaStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.version = source["version"];
	        this.modelReady = source["modelReady"];
	        this.modelName = source["modelName"];
	        this.modelsLoaded = source["modelsLoaded"];
	        this.error = source["error"];
	    }
	}
	
	
	export class QuestUpdatedEvent {
	    completed: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new QuestUpdatedEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.completed = source["completed"];
	        this.count = source["count"];
	    }
	}
	export class RankUpdatedEvent {
	    format: string;
	    tier: string;
	    step: string;
	
	    static createFrom(source: any = {}) {
	        return new RankUpdatedEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.tier = source["tier"];
	        this.step = source["step"];
	    }
	}
	
	export class RecommendationStatsResponse {
	    totalRecommendations: number;
	    acceptedCount: number;
	    rejectedCount: number;
	    ignoredCount: number;
	    alternateCount: number;
	    acceptanceRate: number;
	    acceptancePercent: number;
	    winRateOnAccepted?: number;
	    winRateOnRejected?: number;
	
	    static createFrom(source: any = {}) {
	        return new RecommendationStatsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalRecommendations = source["totalRecommendations"];
	        this.acceptedCount = source["acceptedCount"];
	        this.rejectedCount = source["rejectedCount"];
	        this.ignoredCount = source["ignoredCount"];
	        this.alternateCount = source["alternateCount"];
	        this.acceptanceRate = source["acceptanceRate"];
	        this.acceptancePercent = source["acceptancePercent"];
	        this.winRateOnAccepted = source["winRateOnAccepted"];
	        this.winRateOnRejected = source["winRateOnRejected"];
	    }
	}
	export class RecordActionRequest {
	    recommendationID: string;
	    action: string;
	    alternateChoiceID?: number;
	
	    static createFrom(source: any = {}) {
	        return new RecordActionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.recommendationID = source["recommendationID"];
	        this.action = source["action"];
	        this.alternateChoiceID = source["alternateChoiceID"];
	    }
	}
	export class RecordOutcomeRequest {
	    recommendationID: string;
	    matchID: string;
	    result: string;
	
	    static createFrom(source: any = {}) {
	        return new RecordOutcomeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.recommendationID = source["recommendationID"];
	        this.matchID = source["matchID"];
	        this.result = source["result"];
	    }
	}
	export class RecordRecommendationRequest {
	    recommendationType: string;
	    recommendedCardID?: number;
	    recommendedArchetype?: string;
	    context?: RecommendationContextData;
	    score?: number;
	    rank?: number;
	
	    static createFrom(source: any = {}) {
	        return new RecordRecommendationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.recommendationType = source["recommendationType"];
	        this.recommendedCardID = source["recommendedCardID"];
	        this.recommendedArchetype = source["recommendedArchetype"];
	        this.context = this.convertValues(source["context"], RecommendationContextData);
	        this.score = source["score"];
	        this.rank = source["rank"];
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
	export class RecordRecommendationResponse {
	    recommendationID: string;
	    success: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new RecordRecommendationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.recommendationID = source["recommendationID"];
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class ReplayDraftDetectedEvent {
	    draftId: string;
	    setCode: string;
	    eventType: string;
	
	    static createFrom(source: any = {}) {
	        return new ReplayDraftDetectedEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.draftId = source["draftId"];
	        this.setCode = source["setCode"];
	        this.eventType = source["eventType"];
	    }
	}
	export class ReplayErrorEvent {
	    error: string;
	    code: string;
	    details: string;
	
	    static createFrom(source: any = {}) {
	        return new ReplayErrorEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.error = source["error"];
	        this.code = source["code"];
	        this.details = source["details"];
	    }
	}
	export class ReplayStatus {
	    isActive: boolean;
	    isPaused: boolean;
	    currentEntry: number;
	    totalEntries: number;
	    percentComplete: number;
	    elapsed: number;
	    speed: number;
	    filter: string;
	
	    static createFrom(source: any = {}) {
	        return new ReplayStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isActive = source["isActive"];
	        this.isPaused = source["isPaused"];
	        this.currentEntry = source["currentEntry"];
	        this.totalEntries = source["totalEntries"];
	        this.percentComplete = source["percentComplete"];
	        this.elapsed = source["elapsed"];
	        this.speed = source["speed"];
	        this.filter = source["filter"];
	    }
	}
	
	export class SetInfo {
	    code: string;
	    name: string;
	    iconSvgUri: string;
	    setType: string;
	    releasedAt: string;
	    cardCount: number;
	
	    static createFrom(source: any = {}) {
	        return new SetInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.iconSvgUri = source["iconSvgUri"];
	        this.setType = source["setType"];
	        this.releasedAt = source["releasedAt"];
	        this.cardCount = source["cardCount"];
	    }
	}
	export class StatsUpdatedEvent {
	    matches: number;
	    games: number;
	
	    static createFrom(source: any = {}) {
	        return new StatsUpdatedEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.matches = source["matches"];
	        this.games = source["games"];
	    }
	}
	
	
	

}

export namespace insights {
	
	export class TopCard {
	    name: string;
	    color: string;
	    rarity: string;
	    gihwr: number;
	    cmc?: number;
	
	    static createFrom(source: any = {}) {
	        return new TopCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.color = source["color"];
	        this.rarity = source["rarity"];
	        this.gihwr = source["gihwr"];
	        this.cmc = source["cmc"];
	    }
	}
	export class ArchetypeCards {
	    colors: string;
	    top_cards: TopCard[];
	    top_creatures: TopCard[];
	    top_removal: TopCard[];
	    top_commons: TopCard[];
	
	    static createFrom(source: any = {}) {
	        return new ArchetypeCards(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.colors = source["colors"];
	        this.top_cards = this.convertValues(source["top_cards"], TopCard);
	        this.top_creatures = this.convertValues(source["top_creatures"], TopCard);
	        this.top_removal = this.convertValues(source["top_removal"], TopCard);
	        this.top_commons = this.convertValues(source["top_commons"], TopCard);
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
	export class OverdraftedColor {
	    color: string;
	    win_rate: number;
	    popularity: number;
	    delta: number;
	
	    static createFrom(source: any = {}) {
	        return new OverdraftedColor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.color = source["color"];
	        this.win_rate = source["win_rate"];
	        this.popularity = source["popularity"];
	        this.delta = source["delta"];
	    }
	}
	export class ColorAnalysis {
	    best_mono_color: string;
	    best_color_pair: string;
	    deepest_colors: string[];
	    overdrafted_colors: OverdraftedColor[];
	
	    static createFrom(source: any = {}) {
	        return new ColorAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.best_mono_color = source["best_mono_color"];
	        this.best_color_pair = source["best_color_pair"];
	        this.deepest_colors = source["deepest_colors"];
	        this.overdrafted_colors = this.convertValues(source["overdrafted_colors"], OverdraftedColor);
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
	export class ColorPowerRank {
	    color: string;
	    win_rate: number;
	    games_played: number;
	    popularity: number;
	    rating: string;
	
	    static createFrom(source: any = {}) {
	        return new ColorPowerRank(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.color = source["color"];
	        this.win_rate = source["win_rate"];
	        this.games_played = source["games_played"];
	        this.popularity = source["popularity"];
	        this.rating = source["rating"];
	    }
	}
	export class FormatSpeed {
	    speed: string;
	    avg_game_turn?: number;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new FormatSpeed(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.speed = source["speed"];
	        this.avg_game_turn = source["avg_game_turn"];
	        this.description = source["description"];
	    }
	}
	export class FormatInsights {
	    set_code: string;
	    draft_format: string;
	    color_rankings: ColorPowerRank[];
	    top_bombs: TopCard[];
	    top_removal: TopCard[];
	    top_creatures: TopCard[];
	    top_commons: TopCard[];
	    format_speed: FormatSpeed;
	    color_analysis?: ColorAnalysis;
	
	    static createFrom(source: any = {}) {
	        return new FormatInsights(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.set_code = source["set_code"];
	        this.draft_format = source["draft_format"];
	        this.color_rankings = this.convertValues(source["color_rankings"], ColorPowerRank);
	        this.top_bombs = this.convertValues(source["top_bombs"], TopCard);
	        this.top_removal = this.convertValues(source["top_removal"], TopCard);
	        this.top_creatures = this.convertValues(source["top_creatures"], TopCard);
	        this.top_commons = this.convertValues(source["top_commons"], TopCard);
	        this.format_speed = this.convertValues(source["format_speed"], FormatSpeed);
	        this.color_analysis = this.convertValues(source["color_analysis"], ColorAnalysis);
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

export namespace metrics {
	
	export class LatencyStats {
	    mean: number;
	    p50: number;
	    p95: number;
	    p99: number;
	    min: number;
	    max: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new LatencyStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mean = source["mean"];
	        this.p50 = source["p50"];
	        this.p95 = source["p95"];
	        this.p99 = source["p99"];
	        this.min = source["min"];
	        this.max = source["max"];
	        this.count = source["count"];
	    }
	}
	export class DraftStats {
	    parse_latency: LatencyStats;
	    ratings_latency: LatencyStats;
	    ui_update_latency: LatencyStats;
	    end_to_end_latency: LatencyStats;
	    events_processed: number;
	    packs_rated: number;
	    api_requests: number;
	    api_errors: number;
	    cache_hits: number;
	    cache_misses: number;
	    cache_hit_rate: number;
	    api_success_rate: number;
	    uptime: string;
	
	    static createFrom(source: any = {}) {
	        return new DraftStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parse_latency = this.convertValues(source["parse_latency"], LatencyStats);
	        this.ratings_latency = this.convertValues(source["ratings_latency"], LatencyStats);
	        this.ui_update_latency = this.convertValues(source["ui_update_latency"], LatencyStats);
	        this.end_to_end_latency = this.convertValues(source["end_to_end_latency"], LatencyStats);
	        this.events_processed = source["events_processed"];
	        this.packs_rated = source["packs_rated"];
	        this.api_requests = source["api_requests"];
	        this.api_errors = source["api_errors"];
	        this.cache_hits = source["cache_hits"];
	        this.cache_misses = source["cache_misses"];
	        this.cache_hit_rate = source["cache_hit_rate"];
	        this.api_success_rate = source["api_success_rate"];
	        this.uptime = source["uptime"];
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
	export class Deck {
	    ID: string;
	    AccountID: number;
	    Name: string;
	    Format: string;
	    Description?: string;
	    ColorIdentity?: string;
	    Source: string;
	    DraftEventID?: string;
	    MatchesPlayed: number;
	    MatchesWon: number;
	    GamesPlayed: number;
	    GamesWon: number;
	    CreatedAt: time.Time;
	    ModifiedAt: time.Time;
	    LastPlayed?: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Deck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.AccountID = source["AccountID"];
	        this.Name = source["Name"];
	        this.Format = source["Format"];
	        this.Description = source["Description"];
	        this.ColorIdentity = source["ColorIdentity"];
	        this.Source = source["Source"];
	        this.DraftEventID = source["DraftEventID"];
	        this.MatchesPlayed = source["MatchesPlayed"];
	        this.MatchesWon = source["MatchesWon"];
	        this.GamesPlayed = source["GamesPlayed"];
	        this.GamesWon = source["GamesWon"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], time.Time);
	        this.ModifiedAt = this.convertValues(source["ModifiedAt"], time.Time);
	        this.LastPlayed = this.convertValues(source["LastPlayed"], time.Time);
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
	export class DeckCard {
	    ID: number;
	    DeckID: string;
	    CardID: number;
	    Quantity: number;
	    Board: string;
	    FromDraftPick: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeckCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.DeckID = source["DeckID"];
	        this.CardID = source["CardID"];
	        this.Quantity = source["Quantity"];
	        this.Board = source["Board"];
	        this.FromDraftPick = source["FromDraftPick"];
	    }
	}
	export class DeckMetrics {
	    total_cards: number;
	    total_non_land_cards: number;
	    creature_count: number;
	    noncreature_count: number;
	    cmc_average: number;
	    distribution_all: number[];
	    distribution_creatures: number[];
	    distribution_noncreatures: number[];
	    type_breakdown: Record<string, number>;
	    color_distribution: Record<string, number>;
	    color_counts: Record<string, number>;
	    multi_color_count: number;
	    colorless_count: number;
	
	    static createFrom(source: any = {}) {
	        return new DeckMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_cards = source["total_cards"];
	        this.total_non_land_cards = source["total_non_land_cards"];
	        this.creature_count = source["creature_count"];
	        this.noncreature_count = source["noncreature_count"];
	        this.cmc_average = source["cmc_average"];
	        this.distribution_all = source["distribution_all"];
	        this.distribution_creatures = source["distribution_creatures"];
	        this.distribution_noncreatures = source["distribution_noncreatures"];
	        this.type_breakdown = source["type_breakdown"];
	        this.color_distribution = source["color_distribution"];
	        this.color_counts = source["color_counts"];
	        this.multi_color_count = source["multi_color_count"];
	        this.colorless_count = source["colorless_count"];
	    }
	}
	export class DeckPerformance {
	    DeckID: string;
	    MatchesPlayed: number;
	    MatchesWon: number;
	    MatchesLost: number;
	    GamesPlayed: number;
	    GamesWon: number;
	    GamesLost: number;
	    MatchWinRate: number;
	    GameWinRate: number;
	    LastPlayed?: time.Time;
	    AverageDuration?: number;
	    CurrentWinStreak: number;
	    LongestWinStreak: number;
	    LongestLossStreak: number;
	
	    static createFrom(source: any = {}) {
	        return new DeckPerformance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.DeckID = source["DeckID"];
	        this.MatchesPlayed = source["MatchesPlayed"];
	        this.MatchesWon = source["MatchesWon"];
	        this.MatchesLost = source["MatchesLost"];
	        this.GamesPlayed = source["GamesPlayed"];
	        this.GamesWon = source["GamesWon"];
	        this.GamesLost = source["GamesLost"];
	        this.MatchWinRate = source["MatchWinRate"];
	        this.GameWinRate = source["GameWinRate"];
	        this.LastPlayed = this.convertValues(source["LastPlayed"], time.Time);
	        this.AverageDuration = source["AverageDuration"];
	        this.CurrentWinStreak = source["CurrentWinStreak"];
	        this.LongestWinStreak = source["LongestWinStreak"];
	        this.LongestLossStreak = source["LongestLossStreak"];
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
	export class DeckTag {
	    ID: number;
	    DeckID: string;
	    Tag: string;
	    CreatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new DeckTag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.DeckID = source["DeckID"];
	        this.Tag = source["Tag"];
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
	    OverallGrade?: string;
	    OverallScore?: number;
	    PickQualityScore?: number;
	    ColorDisciplineScore?: number;
	    DeckCompositionScore?: number;
	    StrategicScore?: number;
	    PredictedWinRate?: number;
	    PredictedWinRateMin?: number;
	    PredictedWinRateMax?: number;
	    PredictionFactors?: string;
	    PredictedAt?: time.Time;
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
	        this.OverallGrade = source["OverallGrade"];
	        this.OverallScore = source["OverallScore"];
	        this.PickQualityScore = source["PickQualityScore"];
	        this.ColorDisciplineScore = source["ColorDisciplineScore"];
	        this.DeckCompositionScore = source["DeckCompositionScore"];
	        this.StrategicScore = source["StrategicScore"];
	        this.PredictedWinRate = source["PredictedWinRate"];
	        this.PredictedWinRateMin = source["PredictedWinRateMin"];
	        this.PredictedWinRateMax = source["PredictedWinRateMax"];
	        this.PredictionFactors = source["PredictionFactors"];
	        this.PredictedAt = this.convertValues(source["PredictedAt"], time.Time);
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
	export class Game {
	    ID: number;
	    MatchID: string;
	    GameNumber: number;
	    Result: string;
	    DurationSeconds?: number;
	    ResultReason?: string;
	    CreatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Game(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.MatchID = source["MatchID"];
	        this.GameNumber = source["GameNumber"];
	        this.Result = source["Result"];
	        this.DurationSeconds = source["DurationSeconds"];
	        this.ResultReason = source["ResultReason"];
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
	export class MissingCard {
	    CardID: string;
	    CardName: string;
	    GIHWR: number;
	    Tier: string;
	    PickedAt: number;
	    WheelProbability: number;
	
	    static createFrom(source: any = {}) {
	        return new MissingCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CardID = source["CardID"];
	        this.CardName = source["CardName"];
	        this.GIHWR = source["GIHWR"];
	        this.Tier = source["Tier"];
	        this.PickedAt = source["PickedAt"];
	        this.WheelProbability = source["WheelProbability"];
	    }
	}
	export class MissingCardsAnalysis {
	    SessionID: string;
	    PackNumber: number;
	    PickNumber: number;
	    InitialCards: string[];
	    CurrentCards: string[];
	    PickedByMe: string[];
	    MissingCards: MissingCard[];
	    TotalMissing: number;
	    BombsMissing: number;
	
	    static createFrom(source: any = {}) {
	        return new MissingCardsAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SessionID = source["SessionID"];
	        this.PackNumber = source["PackNumber"];
	        this.PickNumber = source["PickNumber"];
	        this.InitialCards = source["InitialCards"];
	        this.CurrentCards = source["CurrentCards"];
	        this.PickedByMe = source["PickedByMe"];
	        this.MissingCards = this.convertValues(source["MissingCards"], MissingCard);
	        this.TotalMissing = source["TotalMissing"];
	        this.BombsMissing = source["BombsMissing"];
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
	    last_seen_at?: time.Time;
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
	        this.last_seen_at = this.convertValues(source["last_seen_at"], time.Time);
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
	export class RarityCompletion {
	    Rarity: string;
	    Total: number;
	    Owned: number;
	    Percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new RarityCompletion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Rarity = source["Rarity"];
	        this.Total = source["Total"];
	        this.Owned = source["Owned"];
	        this.Percentage = source["Percentage"];
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
	export class SetCompletion {
	    SetCode: string;
	    SetName: string;
	    TotalCards: number;
	    OwnedCards: number;
	    Percentage: number;
	    RarityBreakdown: Record<string, RarityCompletion>;
	
	    static createFrom(source: any = {}) {
	        return new SetCompletion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SetCode = source["SetCode"];
	        this.SetName = source["SetName"];
	        this.TotalCards = source["TotalCards"];
	        this.OwnedCards = source["OwnedCards"];
	        this.Percentage = source["Percentage"];
	        this.RarityBreakdown = this.convertValues(source["RarityBreakdown"], RarityCompletion, true);
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

export namespace prediction {
	
	export class SynergyScore {
	    card_a: string;
	    card_b: string;
	    synergy_type: string;
	    score: number;
	    reason: string;
	    weight: number;
	
	    static createFrom(source: any = {}) {
	        return new SynergyScore(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.card_a = source["card_a"];
	        this.card_b = source["card_b"];
	        this.synergy_type = source["synergy_type"];
	        this.score = source["score"];
	        this.reason = source["reason"];
	        this.weight = source["weight"];
	    }
	}
	export class SynergyResult {
	    overall_score: number;
	    synergy_pairs: SynergyScore[];
	    tribal_synergies: number;
	    mech_synergies: number;
	    color_synergies: number;
	    top_synergies: string[];
	
	    static createFrom(source: any = {}) {
	        return new SynergyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.overall_score = source["overall_score"];
	        this.synergy_pairs = this.convertValues(source["synergy_pairs"], SynergyScore);
	        this.tribal_synergies = source["tribal_synergies"];
	        this.mech_synergies = source["mech_synergies"];
	        this.color_synergies = source["color_synergies"];
	        this.top_synergies = source["top_synergies"];
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
	export class PredictionFactors {
	    deck_average_gihwr: number;
	    color_adjustment: number;
	    curve_score: number;
	    bomb_bonus: number;
	    synergy_score: number;
	    synergy_details?: SynergyResult;
	    baseline_win_rate: number;
	    explanation: string;
	    card_breakdown: Record<string, number>;
	    color_distribution: Record<string, number>;
	    curve_distribution: Record<number, number>;
	    total_cards: number;
	    high_performers: string[];
	    low_performers: string[];
	    confidence_level: string;
	
	    static createFrom(source: any = {}) {
	        return new PredictionFactors(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deck_average_gihwr = source["deck_average_gihwr"];
	        this.color_adjustment = source["color_adjustment"];
	        this.curve_score = source["curve_score"];
	        this.bomb_bonus = source["bomb_bonus"];
	        this.synergy_score = source["synergy_score"];
	        this.synergy_details = this.convertValues(source["synergy_details"], SynergyResult);
	        this.baseline_win_rate = source["baseline_win_rate"];
	        this.explanation = source["explanation"];
	        this.card_breakdown = source["card_breakdown"];
	        this.color_distribution = source["color_distribution"];
	        this.curve_distribution = source["curve_distribution"];
	        this.total_cards = source["total_cards"];
	        this.high_performers = source["high_performers"];
	        this.low_performers = source["low_performers"];
	        this.confidence_level = source["confidence_level"];
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
	export class DeckPrediction {
	    PredictedWinRate: number;
	    PredictedWinRateMin: number;
	    PredictedWinRateMax: number;
	    Factors: PredictionFactors;
	    PredictedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new DeckPrediction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PredictedWinRate = source["PredictedWinRate"];
	        this.PredictedWinRateMin = source["PredictedWinRateMin"];
	        this.PredictedWinRateMax = source["PredictedWinRateMax"];
	        this.Factors = this.convertValues(source["Factors"], PredictionFactors);
	        this.PredictedAt = this.convertValues(source["PredictedAt"], time.Time);
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

