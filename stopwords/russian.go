package stopwords

// Russian contains common Russian stop words to filter out from analysis
var Russian = map[string]bool{
	// Предлоги
	"в": true, "на": true, "с": true, "к": true, "по": true, "за": true,
	"из": true, "у": true, "о": true, "об": true, "от": true, "до": true,
	"для": true, "при": true, "без": true, "под": true, "над": true,
	"про": true, "через": true, "между": true, "перед": true, "после": true,
	"около": true, "возле": true, "вокруг": true, "ради": true,

	// Союзы
	"и": true, "а": true, "но": true, "или": true, "да": true, "ни": true,
	"что": true, "чтобы": true, "если": true, "когда": true, "как": true,
	"потому": true, "поэтому": true, "так": true, "тоже": true, "также": true,
	"хотя": true, "пока": true, "либо": true, "чем": true, "где": true,

	// Местоимения
	"я": true, "ты": true, "он": true, "она": true, "оно": true, "мы": true,
	"вы": true, "они": true, "мне": true, "тебе": true, "ему": true, "ей": true,
	"нам": true, "вам": true, "им": true, "меня": true, "тебя": true, "его": true,
	"её": true, "ее": true, "нас": true, "вас": true, "их": true, "мной": true,
	"тобой": true, "ним": true, "ней": true, "нами": true, "вами": true, "ними": true,
	"себя": true, "себе": true, "собой": true,
	"этот": true, "эта": true, "это": true, "эти": true, "этого": true, "этой": true,
	"этих": true, "этому": true, "этим": true, "этом": true,
	"тот": true, "та": true, "то": true, "те": true, "того": true, "той": true,
	"тех": true, "тому": true, "тем": true, "том": true,
	"кто": true, "какой": true, "какая": true, "какое": true, "какие": true,
	"чей": true, "чья": true, "чьё": true, "чье": true, "чьи": true,
	"который": true, "которая": true, "которое": true, "которые": true,
	"весь": true, "вся": true, "всё": true, "все": true, "всего": true, "всей": true,
	"всех": true, "всему": true, "всем": true,
	"сам": true, "сама": true, "само": true, "сами": true,
	"свой": true, "своя": true, "своё": true, "свое": true, "свои": true,
	"мой": true, "моя": true, "моё": true, "мое": true, "мои": true,
	"твой": true, "твоя": true, "твоё": true, "твое": true, "твои": true,
	"наш": true, "наша": true, "наше": true, "наши": true,
	"ваш": true, "ваша": true, "ваше": true, "ваши": true,

	// Частицы
	"не": true, "бы": true, "же": true, "ли": true, "ведь": true,
	"вот": true, "вон": true, "даже": true, "лишь": true, "только": true,
	"уже": true, "ещё": true, "еще": true, "разве": true, "неужели": true,

	// Наречия
	"очень": true, "там": true, "тут": true, "здесь": true, "туда": true,
	"сюда": true, "оттуда": true, "отсюда": true, "тогда": true,
	"теперь": true, "сейчас": true, "потом": true, "затем": true, "почему": true,
	"зачем": true, "куда": true, "откуда": true, "сколько": true,

	// Глаголы-связки и вспомогательные
	"быть": true, "был": true, "была": true, "было": true, "были": true,
	"буду": true, "будет": true, "будут": true, "будем": true, "будете": true,
	"есть": true, "нет": true, "ну": true,
	"можно": true, "нужно": true, "надо": true, "нельзя": true,

	// Прочие частые слова
	"ага": true, "ок": true, "окей": true, "ладно": true, "хорошо": true,
	"просто": true, "типа": true, "короче": true, "вообще": true, "кстати": true,
	"блин": true, "бля": true, "хз": true, "лол": true, "ахах": true, "хах": true,
	"ахаха": true, "хаха": true, "ахахах": true, "хахаха": true,
	"чё": true, "че": true, "ща": true, "щас": true,
}

// IsStopWord checks if a word is a stop word
func IsStopWord(word string) bool {
	return Russian[word]
}
