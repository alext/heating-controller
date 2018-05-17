package controller_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/thermostat/thermostatfakes"
	"github.com/alext/heating-controller/units"
)

var _ = Describe("Action", func() {

	DescribeTable("string representation",
		func(a controller.Action, str string) {
			Expect(a.String()).To(Equal(str))
		},
		Entry("On", controller.On, "On"),
		Entry("Off", controller.Off, "Off"),
		Entry("SetTarget", controller.SetTarget, "SetTarget"),
		Entry("IncreaseTarget", controller.IncreaseTarget, "IncreaseTarget"),
		Entry("DecreaseTarget", controller.DecreaseTarget, "DecreaseTarget"),
	)

	DescribeTable("JSON marshalling/unmarshalling",
		func(a controller.Action) {
			str := `"` + a.String() + `"`
			Expect(json.Marshal(a)).To(BeEquivalentTo(str))

			var actual controller.Action
			err := json.Unmarshal([]byte(str), &actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(a))
		},
		Entry("On", controller.On),
		Entry("Off", controller.On),
		Entry("SetTarget", controller.SetTarget),
		Entry("IncreaseTarget", controller.IncreaseTarget),
		Entry("DecreaseTarget", controller.DecreaseTarget),
	)

	It("JSON unmarshal should error for an unrecognised string", func() {
		var a controller.Action
		err := json.Unmarshal([]byte(`"Foo"`), &a)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("ThermostatAction", func() {

	type applyCase struct {
		initialTarget units.Temperature
		action        *controller.ThermostatAction
		expectSet     bool
	}

	DescribeTable("applying to a Thermostat",
		func(c applyCase) {
			t := &thermostatfakes.FakeThermostat{}
			t.TargetReturns(c.initialTarget)

			c.action.Apply(t)

			if c.expectSet {
				Expect(t.SetCallCount()).To(Equal(1))
				Expect(t.SetArgsForCall(0)).To(Equal(c.action.Param))
			} else {
				Expect(t.SetCallCount()).To(Equal(0))
			}
		},
		Entry("SetTarget sets the thermostat when lower", applyCase{
			initialTarget: 19500, expectSet: true,
			action: &controller.ThermostatAction{Action: controller.SetTarget, Param: 19000},
		}),
		Entry("SetTarget sets the thermostat when higher", applyCase{
			initialTarget: 17000, expectSet: true,
			action: &controller.ThermostatAction{Action: controller.SetTarget, Param: 19000},
		}),
		Entry("IncreaseTarget increases the thermostat target when it's lower than the param", applyCase{
			initialTarget: 18500, expectSet: true,
			action: &controller.ThermostatAction{Action: controller.IncreaseTarget, Param: 19000},
		}),
		Entry("IncreaseTarget does nothing when the target is higher than the param", applyCase{
			initialTarget: 19500, expectSet: false,
			action: &controller.ThermostatAction{Action: controller.IncreaseTarget, Param: 19000},
		}),
		Entry("DecreaseTarget decreases the thermostat target when it's higher than the param", applyCase{
			initialTarget: 19500, expectSet: true,
			action: &controller.ThermostatAction{Action: controller.DecreaseTarget, Param: 19000},
		}),
		Entry("DecreaseTarget does nothing when the target is lower than the param", applyCase{
			initialTarget: 18500, expectSet: false,
			action: &controller.ThermostatAction{Action: controller.DecreaseTarget, Param: 19000},
		}),
		Entry("does nothing with an unexpected action", applyCase{
			expectSet: false,
			action:    &controller.ThermostatAction{Action: controller.On},
		}),
	)
})
