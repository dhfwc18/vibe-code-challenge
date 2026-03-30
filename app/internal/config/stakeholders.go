package config

// stakeholderSeeds lists all named political figures across the four parties.
// Entry timing controls when a figure becomes available for role assignment.
// See game_design.md Named Cast section for full biographical notes.
var stakeholderSeeds = []StakeholderSeed{

	// ================================================================
	// FARRIGHT -- Taitan First
	// ================================================================

	{
		ID: "joe_barage", Party: PartyFarRight, Role: RoleLeader,
		EntryTiming: TimingStart,
		Name: "Joe Barage",
		Biography: "Born in Murica, moved to Taitan aged 14. Founder of the Taitan for " +
			"Taitons movement. Loud, media-savvy, frames every issue as Taitan sovereignty " +
			"versus foreign interference.",
		IdeologyScore: 95, NetZeroSympathy: 5, RiskTolerance: 85, PopulismScore: 98,
		DiplomaticSkill: 20,
		ConsultancyAffinity: []string{"frontier_energy_institute"},
		Signals: []string{
			"Murica-born, naturalised Taiton",
			"questions the net zero cost to ordinary Taitons",
			"close ties with Murican energy sector donors",
		},
	},
	{
		ID: "rex_harlow", Party: PartyFarRight, Role: RoleChancellor,
		EntryTiming: TimingStart,
		Name: "Rex Harlow",
		Biography: "Old money, former City trader. Sees green spending as economic self-harm.",
		IdeologyScore: 80, NetZeroSympathy: 15, RiskTolerance: 40, PopulismScore: 50,
		DiplomaticSkill: 30,
		ConsultancyAffinity: []string{"albion_institute", "heritage_taitan"},
		Signals: []string{
			"advocate for lower business regulation",
			"opposed the carbon levy bill",
			"strong Taitan financial sector ties",
		},
	},
	{
		ID: "tommy_braveheart", Party: PartyFarRight, Role: RoleForeignSecretary,
		EntryTiming: TimingStart,
		Name: "Thomas \"Tommy\" Braveheart",
		Biography: "Ex-military, no-nonsense. Frames energy security as a sovereignty issue. " +
			"Supports domestic fossil extraction on independence grounds. Will back renewables " +
			"if the case is made on strategic self-sufficiency.",
		IdeologyScore: 70, NetZeroSympathy: 30, RiskTolerance: 60, PopulismScore: 65,
		DiplomaticSkill: 35,
		Signals: []string{
			"decorated military career",
			"Taitan energy independence advocate",
			"pragmatic on low-carbon if framed as security",
		},
	},
	{
		ID: "ticky_tennison", Party: PartyFarRight, Role: RoleEnergy,
		EntryTiming: TimingStart,
		Name: "TD \"Ticky\" Tennison",
		Biography: "Career in Murican-linked energy sector before entering politics. Has " +
			"personal relationships with Murican consultancy and think tank founders.",
		IdeologyScore: 85, NetZeroSympathy: 10, RiskTolerance: 70, PopulismScore: 75,
		DiplomaticSkill: 25,
		Signals: []string{
			"former energy sector lobbyist",
			"regular speaker at Murican energy forums",
			"publicly sceptical of offshore wind economics",
		},
		SpecialMechanic: MechanicTickyPressure,
	},
	{
		ID: "gary_norris", Party: PartyFarRight, Role: RoleChancellor,
		EntryTiming: TimingMid, EntryWeekMin: 350, EntryWeekMax: 550,
		Name: "Gary Norris",
		Biography: "Former Right voter and local councillor who defected after deciding the " +
			"Right had abandoned ordinary working people. The party's sensible face for " +
			"broadcast media. Instinctively anti-net-zero on cost grounds, not ideology.",
		IdeologyScore: 75, NetZeroSympathy: 20, RiskTolerance: 55, PopulismScore: 80,
		DiplomaticSkill: 22,
		Signals: []string{
			"defected from Right citing out-of-touch establishment",
			"talks about energy bills more than sovereignty",
			"surprisingly good on local radio",
		},
	},

	// ================================================================
	// LEFT -- Taitan Labour equivalent
	// ================================================================

	{
		ID: "jj_cameron", Party: PartyLeft, Role: RoleLeader,
		EntryTiming: TimingStart,
		Name: "JJ Cameron",
		Biography: "Veteran backbencher. Former secondary school teacher (RE and History), " +
			"lifelong trade union rep, self-published author of seven pamphlets including " +
			"The Carbon Commons: Energy as a Public Good (2003), Manhole Covers of the " +
			"Taitan Waterboard District (2006, collector's edition), and Why I Was Right: " +
			"A Retrospective (2019, unsolicited). Keeps an allotment. Cycles everywhere. " +
			"Has been on the record about net zero and energy justice since 2001. Replies " +
			"promptly to all correspondence.",
		IdeologyScore: -78, NetZeroSympathy: 87, RiskTolerance: 38, PopulismScore: 72,
		DiplomaticSkill: 30,
		Signals: []string{
			"been saying this since 2001",
			"enormous membership support, limited swing-voter appeal",
			"famously neutral on the Murican defence alliance question",
			"allotment open to the public every second Sunday",
			"prefers academic evidence to consultancy reports",
		},
		SpecialMechanic: MechanicElectoralFatigue,
	},
	{
		// Successor leader; only enters pool after JJ Cameron departs via ELECTORAL_FATIGUE
		ID: "david_reeve", Party: PartyLeft, Role: RoleLeader,
		EntryTiming: TimingSuccessor,
		Name: "David Reeve",
		Biography: "A sensible figure with broadly correct instincts. Has views on things. " +
			"Considered reliable. The party's attempt to appear electable to people who " +
			"did not vote for it last time.",
		IdeologyScore: -25, NetZeroSympathy: 65, RiskTolerance: 45, PopulismScore: 40,
		DiplomaticSkill: 48,
		Signals: []string{
			"modernising figure within the party",
			"backed the Climate Commitment Act",
			"known for consensus-building style",
		},
	},
	{
		ID: "george_harmon", Party: PartyLeft, Role: RoleChancellor,
		EntryTiming: TimingStart,
		Name: "George Harmon",
		Biography: "Fiscal conservative in a left-of-centre mould. Supportive of green " +
			"investment when the numbers work. Will push back on uncosted policies.",
		IdeologyScore: -15, NetZeroSympathy: 55, RiskTolerance: 30, PopulismScore: 25,
		DiplomaticSkill: 40,
		ConsultancyAffinity: []string{"progress_policy_centre"},
		Signals: []string{
			"former public sector economist",
			"authored the Green Investment Framework",
			"reputation for asking where the money is coming from",
			"respects centre-left policy research",
		},
	},
	{
		ID: "john_ashworth", Party: PartyLeft, Role: RoleForeignSecretary,
		EntryTiming: TimingStart,
		Name: "John Ashworth",
		Biography: "Experienced. Reliable. Has strong feelings about correct procedure. " +
			"Represents the solid, unremarkable competence the party relies on.",
		IdeologyScore: -20, NetZeroSympathy: 45, RiskTolerance: 50, PopulismScore: 35,
		DiplomaticSkill: 62,
		Signals: []string{
			"experienced diplomat and parliamentarian",
			"cautious multilateralist",
			"strong feelings about correct procedure",
		},
	},
	{
		ID: "claire_blackwell", Party: PartyLeft, Role: RoleEnergy,
		EntryTiming: TimingStart,
		Name: "Claire Blackwell",
		Biography: "Player's first boss. Left governs at game start in 2010. Committed to " +
			"getting things done within the system. Good starting relationship with player.",
		IdeologyScore: -30, NetZeroSympathy: 70, RiskTolerance: 40, PopulismScore: 30,
		DiplomaticSkill: 42,
		Signals: []string{
			"backed the Renewable Obligation scheme",
			"known for careful stakeholder management",
			"willing to consider nuclear as part of the mix",
		},
	},
	{
		ID: "nia_okafor", Party: PartyLeft, Role: RoleChancellor,
		EntryTiming: TimingMid, EntryWeekMin: 300, EntryWeekMax: 520,
		Name: "Nia Okafor",
		Biography: "Second-generation Taitan, parents from West Afria. Grew up in a " +
			"post-industrial northern constituency on the wrong end of every energy price " +
			"spike since 1990. Her fiscal instincts are shaped by that history.",
		IdeologyScore: -18, NetZeroSympathy: 62, RiskTolerance: 28, PopulismScore: 30,
		DiplomaticSkill: 44,
		Signals: []string{
			"former regional development economist",
			"co-authored the Green Investment Framework",
			"reputation for asking where the money is coming from and why",
		},
	},
	{
		ID: "tariq_rashid", Party: PartyLeft, Role: RoleForeignSecretary,
		EntryTiming: TimingMid, EntryWeekMin: 320, EntryWeekMax: 560,
		Name: "Tariq Rashid",
		Biography: "Third-generation Taitan. Former career diplomat who specialised in " +
			"multilateral environment and trade agreements before returning to domestic " +
			"politics through the party's community organising wing.",
		IdeologyScore: -22, NetZeroSympathy: 48, RiskTolerance: 48, PopulismScore: 33,
		DiplomaticSkill: 78,
		Signals: []string{
			"former career diplomat, environment and trade specialisation",
			"coalition-minded, earns trust through preparation",
			"effective in multilateral forums",
		},
	},
	{
		ID: "marcus_osei", Party: PartyLeft, Role: RoleEnergy,
		EntryTiming: TimingLate, EntryWeekMin: 680, EntryWeekMax: 900,
		Name: "Marcus Osei",
		Biography: "Third-generation Taitan. Grew up in a port city, first in his family " +
			"to attend university. Came up through community organising and trade union work. " +
			"Holds a structural view of the energy transition linking fuel poverty, industrial " +
			"policy, and international climate finance as parts of the same problem.",
		IdeologyScore: -40, NetZeroSympathy: 85, RiskTolerance: 55, PopulismScore: 60,
		DiplomaticSkill: 50,
		Signals: []string{
			"founding member of the Climate Justice Caucus",
			"came up through community organising",
			"genuinely exciting to listen to, which makes the party nervous",
		},
	},

	// ================================================================
	// RIGHT -- Taitan Conservative equivalent
	// ================================================================

	{
		ID: "daniel_cavendish", Party: PartyRight, Role: RoleLeader,
		EntryTiming: TimingStart,
		Name: "Daniel \"Pork\" Cavendish",
		Nickname: "Pork",
		Biography: "Old-money upper class, went to the right schools, inherited the right " +
			"connections. Somewhat flaky under pressure. Privately supportive of net zero " +
			"but easily swayed by party faction pressure. The exact origin of his nickname " +
			"is a mystery: associates deflect when asked, and he simply smiles.",
		IdeologyScore: 35, NetZeroSympathy: 60, RiskTolerance: 30, PopulismScore: 25,
		DiplomaticSkill: 50,
		ConsultancyAffinity: []string{"meridian_strategy"},
		Signals: []string{
			"educated at Briarfield and Aldenvale",
			"considered a safe pair of hands by the establishment",
			"occasionally says the right thing for unclear reasons",
			"known for expensive consultancy commissions",
		},
	},
	{
		ID: "philip_drake", Party: PartyRight, Role: RoleChancellor,
		EntryTiming: TimingStart,
		Name: "Philip Drake",
		Biography: "Free market purist, deeply sceptical of green industrial policy. " +
			"Describes every subsidy as market distortion and means it.",
		IdeologyScore: 55, NetZeroSympathy: 35, RiskTolerance: 35, PopulismScore: 30,
		DiplomaticSkill: 38,
		ConsultancyAffinity: []string{"albion_institute", "energy_realists_network"},
		Signals: []string{
			"authored free market think tank papers on energy",
			"opposed the windfall tax",
			"wants to scrap net zero subsidies",
		},
	},
	{
		ID: "andrew_stafford", Party: PartyRight, Role: RoleForeignSecretary,
		EntryTiming: TimingStart,
		Name: "Andrew Stafford",
		Biography: "Hawkish, frames energy security as strategic priority. Can be won over " +
			"on renewables if the argument is made in terms of reducing Taitan dependence " +
			"on foreign fossil fuel states.",
		IdeologyScore: 50, NetZeroSympathy: 40, RiskTolerance: 60, PopulismScore: 40,
		DiplomaticSkill: 55,
		Signals: []string{
			"hawkish on foreign policy",
			"frames energy as a strategic dependency issue",
			"open to renewables on security grounds",
		},
	},
	{
		ID: "rupert_holm", Party: PartyRight, Role: RoleEnergy,
		EntryTiming: TimingStart,
		Name: "Rupert Holm",
		Biography: "Tech-optimist, strongly pro-nuclear, lukewarm on mandates and home " +
			"retrofit. Will approve technology-led policies readily; resists behaviour-change " +
			"policies.",
		IdeologyScore: 45, NetZeroSympathy: 60, RiskTolerance: 55, PopulismScore: 35,
		DiplomaticSkill: 35,
		ConsultancyAffinity: []string{"clearpath_advisory", "tacute_energy"},
		Signals: []string{
			"backed the new nuclear programme",
			"sceptical of heat pump mandates",
			"wants to cut planning red tape for renewables",
			"strong working relationships with established energy consultancies",
		},
	},
	{
		ID: "dawn_truscott", Party: PartyRight, Role: RoleChancellor,
		EntryTiming: TimingMid, EntryWeekMin: 260, EntryWeekMax: 460,
		Name: "Dawn \"Dizzy\" Truscott",
		Nickname: "Dizzy",
		Biography: "Sharp, ideologically driven free-marketeer. Believes the state should " +
			"step back and let capital decarbonise on its own schedule. High personal ambition.",
		IdeologyScore: 60, NetZeroSympathy: 30, RiskTolerance: 75, PopulismScore: 40,
		DiplomaticSkill: 35,
		ConsultancyAffinity: []string{"meridian_strategy", "axiom_infrastructure"},
		Signals: []string{
			"youngest ever Chancellor candidate",
			"authored the Truscott Compact on fiscal rules",
			"close ties with Taitan financial sector and Meridian Strategy alumni",
		},
		SpecialMechanic: MechanicDizzySurge,
	},
	{
		ID: "noris_jackson", Party: PartyRight, Role: RoleForeignSecretary,
		EntryTiming: TimingMid, EntryWeekMin: 60, EntryWeekMax: 80,
		Name: "Noris Jackson",
		Biography: "Born in Murica while his father was completing a fellowship there; " +
			"grew up in Taitan from age four. Part of his paternal lineage traces back to " +
			"the Eastern Provinces. Educated at Briarfield and Aldenvale. Spent several " +
			"years as a political journalist and foreign correspondent before entering " +
			"politics, giving him a talent for a good story and a relaxed relationship " +
			"with factual precision. Laddish, loud, socially well-connected. Holds dual " +
			"Murican-Taitan citizenship, which Barage has occasionally used against him.",
		IdeologyScore: 40, NetZeroSympathy: 45, RiskTolerance: 65, PopulismScore: 60,
		DiplomaticSkill: 72,
		Signals: []string{
			"born in Murica, dual citizen",
			"Briarfield and Aldenvale, former foreign correspondent",
			"known for after-hours diplomatic back-channels",
			"never misses a party conference drinks reception",
		},
	},
	{
		ID: "ajay_mehta", Party: PartyRight, Role: RoleEnergy,
		EntryTiming: TimingLate, EntryWeekMin: 650, EntryWeekMax: 850,
		Name: "Ajay Mehta",
		Biography: "Second-generation Taitan, parents from East Azaria. Grew up in a nuclear " +
			"engineering household. The family dinner table consensus was that fission is the " +
			"answer and everything else is sentiment. He has not revised this.",
		IdeologyScore: 45, NetZeroSympathy: 62, RiskTolerance: 55, PopulismScore: 35,
		DiplomaticSkill: 38,
		ConsultancyAffinity: []string{"clearpath_advisory", "tacute_energy"},
		Signals: []string{
			"backed the new nuclear programme",
			"sceptical of heat pump mandates",
			"publicly dismissive of behaviour-change campaigns",
		},
	},
	{
		ID: "sandra_obi_williams", Party: PartyRight, Role: RoleEnergy,
		EntryTiming: TimingLate, EntryWeekMin: 700, EntryWeekMax: 950,
		Name: "Sandra Obi-Williams",
		Biography: "Grew up in a Right-leaning coastal town. Modernising Conservative who " +
			"genuinely believes the market can drive net zero faster than state intervention. " +
			"Her climate record is real but her framing is market-first.",
		IdeologyScore: 38, NetZeroSympathy: 65, RiskTolerance: 45, PopulismScore: 42,
		DiplomaticSkill: 50,
		Signals: []string{
			"youngest Right MP in her intake",
			"backed the Green Innovation Fund",
			"vocal critic of anti-growth climate pessimism",
		},
	},

	// ================================================================
	// FARLEFT -- Taitan Progressive Alliance
	// ================================================================

	{
		ID: "miriam_corbett", Party: PartyFarLeft, Role: RoleLeader,
		EntryTiming: TimingStart,
		Name: "Miriam Corbett",
		Biography: "Long-standing socialist. Strong climate action but only via public " +
			"ownership. Hostile to corporate LCT companies and market mechanisms.",
		IdeologyScore: -90, NetZeroSympathy: 80, RiskTolerance: 55, PopulismScore: 70,
		DiplomaticSkill: 35,
		ConsultancyAversion: true,
		Signals: []string{
			"advocate for nationalised energy",
			"opposes carbon trading as greenwash",
			"strong trade union backing",
			"hostile to private consultancy spend",
		},
	},
	{
		ID: "priya_sharma", Party: PartyFarLeft, Role: RoleChancellor,
		EntryTiming: TimingStart,
		Name: "Priya Sharma",
		Biography: "Academic economist, MMT-leaning. Believes public investment can fund the " +
			"transition without fiscal constraints. Opposed to austerity framing.",
		IdeologyScore: -80, NetZeroSympathy: 75, RiskTolerance: 65, PopulismScore: 50,
		DiplomaticSkill: 48,
		ConsultancyAversion: true,
		Signals: []string{
			"authored The Green New Deal for Taitan",
			"favours debt-funded transition",
			"hostile to private consultancy spend",
		},
	},
	{
		ID: "marcus_webb", Party: PartyFarLeft, Role: RoleForeignSecretary,
		EntryTiming: TimingStart,
		Name: "Marcus Webb",
		Biography: "Ex-military turned pacifist MP. Advocates redirecting defence budget to " +
			"climate diplomacy and green aid. Deeply uncomfortable with hard-power responses.",
		IdeologyScore: -75, NetZeroSympathy: 70, RiskTolerance: 45, PopulismScore: 55,
		DiplomaticSkill: 55,
		Signals: []string{
			"ex-military turned pacifist",
			"advocates redirecting defence budget to climate diplomacy",
			"de-escalatory instinct in all international situations",
		},
	},
	{
		ID: "rosa_chen", Party: PartyFarLeft, Role: RoleEnergy,
		EntryTiming: TimingStart,
		Name: "Rosa Chen",
		Biography: "Strong climate champion, anti-corporate. Will approve ambitious policies " +
			"rapidly but demands they exclude private profit. Hostile to Tacute and private " +
			"consultancies.",
		IdeologyScore: -85, NetZeroSympathy: 95, RiskTolerance: 70, PopulismScore: 65,
		DiplomaticSkill: 40,
		ConsultancyAversion: true,
		Signals: []string{
			"co-authored the Zero Carbon Cities bill",
			"publicly attacked Meridian Strategy for profiteering from climate action",
			"prefers academic evidence over consultancy",
		},
	},
	{
		ID: "declan_murphy", Party: PartyFarLeft, Role: RoleChancellor,
		EntryTiming: TimingMid, EntryWeekMin: 380, EntryWeekMax: 580,
		Name: "Declan Murphy",
		Biography: "Irish-Taitan heritage, community energy cooperative organiser from a " +
			"former mining constituency. Pragmatic within the FarLeft framework: wants public " +
			"ownership but also wants the lights to stay on.",
		IdeologyScore: -70, NetZeroSympathy: 78, RiskTolerance: 50, PopulismScore: 68,
		DiplomaticSkill: 42,
		Signals: []string{
			"founded the Taitan Community Energy Network",
			"known for getting things built",
			"occasionally describes Rosa Chen as impractical",
		},
	},
	{
		ID: "amara_diallo", Party: PartyFarLeft, Role: RoleEnergy,
		EntryTiming: TimingLate, EntryWeekMin: 750, EntryWeekMax: 1000,
		Name: "Amara Diallo",
		Biography: "Mixed Taitan-West Afrian heritage. Grew up in a coastal city watching " +
			"the climate change in real time -- flood lines creeping up the seawall every " +
			"decade. Climate justice framing is not theoretical for her. Youngest figure in " +
			"the FarLeft pool.",
		IdeologyScore: -82, NetZeroSympathy: 98, RiskTolerance: 75, PopulismScore: 72,
		DiplomaticSkill: 52,
		Signals: []string{
			"youngest FarLeft MP ever elected",
			"testified before the Taitan Climate Committee at 19",
			"describes net zero by 2050 as not fast enough",
		},
	},
}
