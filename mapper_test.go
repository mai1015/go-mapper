package go_mapper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type SourceTypeA struct {
	Foo int
	Bar string
}

type DestTypeA struct {
	Foo int
	Bar string
}

type DestTypeBNest struct {
	Bar string
}

type TargetTypeANest struct {
	Foo int
	DestTypeBNest
}

func TestMap(t *testing.T) {
	sourceA, targetA := SourceTypeA{
		1,
		"bar",
	}, DestTypeA{}
	Map(sourceA, &targetA, false)
	assert.Equal(t, sourceA.Foo, targetA.Foo, "did not map Foo")
	assert.Equal(t, sourceA.Bar, targetA.Bar, "did not map Bar")

	targetB := DestTypeBNest{}
	Map(sourceA, &targetB, true)
	assert.Equal(t, sourceA.Bar, targetB.Bar, "did not map Bar")
}

func TestMapWithNest(t *testing.T) {
	sourceA, targetA := SourceTypeA{
		1,
		"bar",
	}, TargetTypeANest{}
	Map(sourceA, &targetA, true)
	assert.Equal(t, sourceA.Foo, targetA.Foo, "did not map Foo")
	assert.Equal(t, sourceA.Bar, targetA.Bar, "did not map Bar")

	sourceB, targetB := TargetTypeANest{
		1,
		DestTypeBNest{
			"bar",
		},
	}, SourceTypeA{}
	Map(sourceB, &targetB, true)
	assert.Equal(t, sourceA.Foo, targetA.Foo, "did not map Foo")
	assert.Equal(t, sourceA.Bar, targetA.Bar, "did not map Bar")
}

type TimeWrapper struct {
	T time.Time
}

type SourceTypeCWrapper struct {
	Time TimeWrapper
}

type SourceTypeC struct {
	Time time.Time
}

type TargetTypeC struct {
	Time time.Time
}

func TestMapWithStruct(t *testing.T) {
	sourceA, targetA := SourceTypeC{
		time.Now(),
	}, TargetTypeC{}

	Map(sourceA, &targetA, false)
	assert.Equal(t, sourceA.Time, targetA.Time, "did not map time")
}

func TestMapWithWrapper(t *testing.T) {
	SetWrapperType("go_mapper.TimeWrapper", true)

	sourceA, targetA := SourceTypeCWrapper{
		TimeWrapper{T: time.Now()},
	}, TargetTypeC{}

	Map(sourceA, &targetA, false)
	assert.Equal(t, sourceA.Time.T, targetA.Time, "did not map time")
}

func TestPanicWhenDestIsNotPointer(t *testing.T) {
	defer func() { recover() }()
	source, dest := SourceTypeA{}, DestTypeA{}
	Map(source, dest, false)

	t.Error("Should have panicked")
}

func TestDestinationIsUpdatedFromSource(t *testing.T) {
	source, dest := SourceTypeA{Foo: 42}, DestTypeA{}
	Map(source, &dest, false)
	assert.Equal(t, 42, dest.Foo)
}

func TestDestinationIsUpdatedFromSourceWhenSourcePassedAsPtr(t *testing.T) {
	source, dest := SourceTypeA{42, "Bar"}, DestTypeA{}
	Map(&source, &dest, false)
	assert.Equal(t, 42, dest.Foo)
	assert.Equal(t, "Bar", dest.Bar)
}

func TestWithNestedTypes(t *testing.T) {
	source := struct {
		Baz   string
		Child SourceTypeA
	}{}
	dest := struct {
		Baz   string
		Child DestTypeA
	}{}

	source.Baz = "Baz"
	source.Child.Bar = "Bar"
	Map(&source, &dest, false)
	assert.Equal(t, "Baz", dest.Baz)
	assert.Equal(t, "Bar", dest.Child.Bar)
}

func TestWithSourceSecondLevel(t *testing.T) {
	source := struct {
		Child DestTypeA
	}{}
	dest := SourceTypeA{}

	source.Child.Bar = "Bar"
	Map(&source, &dest, false)
	assert.Equal(t, "Bar", dest.Bar)
}

func TestWithDestSecondLevel(t *testing.T) {
	source := SourceTypeA{}
	dest := struct {
		Child DestTypeA
	}{}

	source.Bar = "Bar"
	Map(&source, &dest, false)
	assert.Equal(t, "Bar", dest.Child.Bar)
}

func TestWithSliceTypes(t *testing.T) {
	source := struct {
		Children []SourceTypeA
	}{}
	dest := struct {
		Children []DestTypeA
	}{}
	source.Children = []SourceTypeA{
		SourceTypeA{Foo: 1},
		SourceTypeA{Foo: 2}}

	Map(&source, &dest, false)
	assert.Equal(t, 1, dest.Children[0].Foo)
	assert.Equal(t, 2, dest.Children[1].Foo)
}

func TestWithMultiLevelSlices(t *testing.T) {
	source := struct {
		Parents []SourceParent
	}{}
	dest := struct {
		Parents []DestParent
	}{}
	source.Parents = []SourceParent{
		SourceParent{
			Children: []SourceTypeA{
				SourceTypeA{Foo: 42},
				SourceTypeA{Foo: 43},
			},
		},
		SourceParent{
			Children: []SourceTypeA{},
		},
	}

	Map(&source, &dest, false)
}

func TestWithEmptySliceAndIncompatibleTypes(t *testing.T) {
	defer func() { recover() }()

	source := struct {
		Children []struct{ Foo string }
	}{}
	dest := struct {
		Children []struct{ Bar int }
	}{}

	Map(&source, &dest, false)
	t.Error("Should have panicked")
}

func TestWhenSourceIsMissingField(t *testing.T) {
	defer func() { recover() }()
	source := struct {
		A string
	}{}
	dest := struct {
		A, B string
	}{}
	Map(&source, &dest, false)
	t.Error("Should have panicked")
}

func TestWithUnnamedFields(t *testing.T) {
	source := struct {
		Baz string
		SourceTypeA
	}{}
	dest := struct {
		Baz string
		DestTypeA
	}{}
	source.Baz = "Baz"
	source.SourceTypeA.Foo = 42

	Map(&source, &dest, false)
	assert.Equal(t, "Baz", dest.Baz)
	assert.Equal(t, 42, dest.DestTypeA.Foo)
}

func TestWithPointerFieldsNotNil(t *testing.T) {
	source := struct {
		Foo *SourceTypeA
	}{}
	dest := struct {
		Foo *DestTypeA
	}{}
	source.Foo = nil

	Map(&source, &dest, false)
	assert.Nil(t, dest.Foo)
}

func TestWithPointerFieldsNil(t *testing.T) {
	source := struct {
		Foo *SourceTypeA
	}{}
	dest := struct {
		Foo *DestTypeA
	}{}
	source.Foo = &SourceTypeA{Foo: 42}

	Map(&source, &dest, false)
	assert.NotNil(t, dest.Foo)
	assert.Equal(t, 42, dest.Foo.Foo)
}

func TestMapFromPointerToNonPointerTypeWithData(t *testing.T) {
	source := struct {
		Foo *SourceTypeA
	}{}
	dest := struct {
		Foo DestTypeA
	}{}
	source.Foo = &SourceTypeA{Foo: 42}

	Map(&source, &dest, false)
	assert.NotNil(t, dest.Foo)
	assert.Equal(t, 42, dest.Foo.Foo)
}

func TestMapFromPointerToNonPointerTypeWithoutData(t *testing.T) {
	source := struct {
		Foo *SourceTypeA
	}{}
	dest := struct {
		Foo DestTypeA
	}{}
	source.Foo = nil

	Map(&source, &dest, false)
	assert.NotNil(t, dest.Foo)
	assert.Equal(t, 0, dest.Foo.Foo)
}

func TestMapFromPointerToAnonymousTypeToFieldName(t *testing.T) {
	source := struct {
		*SourceTypeA
	}{}
	dest := struct {
		Foo int
	}{}
	source.SourceTypeA = nil

	Map(&source, &dest, false)
	assert.Equal(t, 0, dest.Foo)
}

func TestMapFromPointerToNonPointerTypeWithoutDataAndIncompatibleType(t *testing.T) {
	defer func() { recover() }()
	// Just make sure we stil panic
	source := struct {
		Foo *SourceTypeA
	}{}
	dest := struct {
		Foo struct {
			Baz string
		}
	}{}
	source.Foo = nil

	Map(&source, &dest, false)
	t.Error("Should have panicked")
}

func TestWhenUsingIncompatibleTypes(t *testing.T) {
	defer func() { recover() }()
	source := struct{ Foo string }{}
	dest := struct{ Foo int }{}
	Map(&source, &dest, false)
	t.Error("Should have panicked")
}

func TestWithLooseOption(t *testing.T) {
	source := struct {
		Foo string
		Baz int
	}{"Foo", 42}
	dest := struct {
		Foo string
		Bar int
	}{}
	Map(&source, &dest, true)
	assert.Equal(t, dest.Foo, "Foo")
	assert.Equal(t, dest.Bar, 0)
}

type SourceParent struct {
	Children []SourceTypeA
}

type DestParent struct {
	Children []DestTypeA
}

type SourceTypeB struct {
	A SourceTypeA
	B SourceTypeA
}

type DestTypeB struct {
	A DestTypeA
	B SourceTypeA
}

func TestMapWithSameType(t *testing.T) {
	source := SourceTypeB{
		SourceTypeA{
			1,
			"test1",
		},
		SourceTypeA{
			2,
			"test2",
		},
	}
	dest := DestTypeB{}

	Map(source, &dest, false)
	assert.Equal(t, source.A.Foo, dest.A.Foo, "cannot map bar")
	assert.Equal(t, source.A.Bar, dest.A.Bar, "cannot map bar")
	assert.Equal(t, source.B.Foo, dest.B.Foo, "cannot map bar")
	assert.Equal(t, source.B.Bar, dest.B.Bar, "cannot map bar")
}

type SourceTypeBPtr struct {
	A SourceTypeA
	B *SourceTypeA
}

type DestTypeBPtr struct {
	A DestTypeA
	B *SourceTypeA
}

func TestMapWithSameDestTypePtr(t *testing.T) {
	source := SourceTypeB{
		SourceTypeA{
			1,
			"test1",
		},
		SourceTypeA{
			2,
			"test2",
		},
	}
	dest := DestTypeBPtr{}

	Map(source, &dest, false)
	assert.Equal(t, source.A.Foo, dest.A.Foo, "cannot map bar")
	assert.Equal(t, source.A.Bar, dest.A.Bar, "cannot map bar")
	assert.Equal(t, source.B.Foo, dest.B.Foo, "cannot map bar")
	assert.Equal(t, source.B.Bar, dest.B.Bar, "cannot map bar")
}

func TestMapWithSameSourceTypePtr(t *testing.T) {
	source := SourceTypeBPtr{
		SourceTypeA{
			1,
			"test1",
		},
		&SourceTypeA{
			2,
			"test2",
		},
	}
	dest := DestTypeB{}

	Map(source, &dest, false)
	assert.Equal(t, source.A.Foo, dest.A.Foo, "cannot map bar")
	assert.Equal(t, source.A.Bar, dest.A.Bar, "cannot map bar")
	assert.Equal(t, source.B.Foo, dest.B.Foo, "cannot map bar")
	assert.Equal(t, source.B.Bar, dest.B.Bar, "cannot map bar")
}

func TestMapWithSameTypeBothPtr(t *testing.T) {
	source := SourceTypeBPtr{
		SourceTypeA{
			1,
			"test1",
		},
		&SourceTypeA{
			2,
			"test2",
		},
	}
	dest := DestTypeBPtr{}

	Map(source, &dest, false)
	assert.Equal(t, source.A.Foo, dest.A.Foo, "cannot map bar")
	assert.Equal(t, source.A.Bar, dest.A.Bar, "cannot map bar")
	assert.Equal(t, source.B.Foo, dest.B.Foo, "cannot map bar")
	assert.Equal(t, source.B.Bar, dest.B.Bar, "cannot map bar")
}
