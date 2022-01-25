package snippet

// https://phlatphrog.medium.com/one-way-to-do-it-six-variations-cd58602ac06d
// 1. origin
// 开车去商店
shopper, err := shopper.Drive(FuelNeededToGetToStore)
if nil != err {
	log.Fatalf("could not complete shopping: %s", err)
}
// 买鸡蛋
shopper, err = shopper.BuyEggs(EggsRequired)
if nil != err {
	log.Fatalf("could not complete shopping: %s", err)
}
// 买完鸡蛋开车回家
shopper, err = shopper.Drive(FuelNeededToGetHome)
if nil != err {
	log.Fatalf("could not complete shopping: %s", err)
}

// refact 1: err集中判断
func main(){
	shopper, err := shopper.Drive(FuelNeededToGetToStore)
	FatalIfErrNotNil(err)
	shopper, err = shopper.BuyEggs(EggsRequired)
	FatalIfErrNotNil(err)
	shopper, err = shopper.Drive(FuelNeededToGetHome)
	FatalIfErrNotNil(err)
}

func FatalIfErrNotNil(err error) {
	if nil != err {
		log.Fatalf("could not complete shopping: %s", err)
	}
}

// or
func main(){
	shopper, err := shopper.Drive(FuelNeededToGetToStore)
	if ErrorHandled(err) {
		return ...
	}
}

func ErrorHandled(err) bool {
	if nil != err {
		return true
	}
	// 也可以对error进行其他判断操作

	return false
}

// refact 3: 方法中包括err处理
func main(){
	shopper, err := shopper.Drive(FuelNeededToGetToStore, nil)
	shopper, err = shopper.BuyEggs(EggsRequired, err)
	shopper, err = shopper.Drive(FuelNeededToGetHome, err)
	if nil != err {
		log.Fatalf("could not complete shopping: %s", err)
	}
}

func (s Shopper) Drive(fuelRequired int, err error) (Shopper, error) {
	if nil != err {
		return s, err
	}
	// 业务逻辑处理

	return s, nil
}

// refact 4: 使用函数对err分解
func main(){
	drive := ErrCheckFunc(Drive)
	buy := ErrCheckFunc(BuyEggs)
	err, shopper := drive(nil, shopper, FuelNeededToGetToStore)
	err, shopper = buy(err, shopper, EggsRequired)
	err, shopper = drive(err, shopper, FuelNeededToGetHome)
	if nil != err {
		log.Fatalf("could not complete shopping: %s", err)
	}
}

func Drive(s Shopper, fuelRequired int) (Shopper, error) {
	if nil != err {
		return s, err
	}

	return s, nil
}

func ErrCheckFunc(f func(Shopper, int) (Shopper, error)) func(error, Shopper, int) (error, Shopper) {
	return func(err error, s Shopper, arg int) (error, Shopper) {
		if nil != err {
			return err, s
		}
		s, err = f(s, arg)
		return err, s
	}
}


// refact 5: 单行操作
func main(){
	driveToStore := ErrCheckFunc(Drive, FuelNeededToGetToStore)
	buyEggs := ErrCheckFunc(BuyEggs, EggsRequired)
	driveHome := ErrCheckFunc(Drive, FuelNeededToGetHome)

	err, shopper := driveHome(buyEggs(driveToStore(nil, shopper)))
	if nil != err {
		log.Fatalf("could not complete shopping: %s", err)
	}
}

func ErrCheckFunc(f func(Shopper, int) (Shopper, error), arg int) func(error, Shopper) (error, Shopper) {
	return func(err error, s Shopper) (error, Shopper) {
		if nil != err {
			return err, s
		}
		s, err = f(s, arg)
		return err, s
	}
}

// refact 6: 迭代变换
func main(){
	driveToStore := Flavor(Drive, FuelNeededToGetToStore)
	buyEggs := Flavor(BuyEggs, EggsRequired)
	driveHome := Flavor(Drive, FuelNeededToGetHome)

	shopper, err := ProcessSteps(shopper,
		driveToStore,
		buyEggs,
		driveHome,
	)
	if nil != err {
		log.Fatalf("could not complete shopping: %s", err)
	}
}

func ProcessSteps(s Shopper, steps ...func(Shopper) (Shopper, error)) (Shopper, error) {
	for _, step := range steps {
		var err error
		s, err = step(s)
		if nil != err {
			return s, err
		}
	}
	return s, nil
}

func Flavor(f func(Shopper, int) (Shopper, error), arg int) func(Shopper) (Shopper, error) {
	return func(s Shopper) (Shopper, error) {
		return f(s, arg)
	}
}

