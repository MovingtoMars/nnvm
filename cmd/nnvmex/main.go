package main

import (
	"fmt"
	"log"
	"os"

	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/ssa/analysis"
	"github.com/MovingtoMars/nnvm/ssa/validate"
	"github.com/MovingtoMars/nnvm/target/amd64"
	"github.com/MovingtoMars/nnvm/target/platform"
	"github.com/MovingtoMars/nnvm/types"
)

func main() {
	test()
}

func test() {
	mod := ssa.NewModule("mod")

	strLit := ssa.NewStringLiteral("%2d: %lld\n", true)
	strGlob := mod.NewGlobal(strLit.Type(), ssa.NewLiteralInitialiser(strLit), "str")

	failLit := ssa.NewStringLiteral("Please supply number as argument\n", true)
	failGlob := mod.NewGlobal(failLit.Type(), ssa.NewLiteralInitialiser(failLit), "str")

	i8ptr := types.NewPointer(types.NewInt(8))
	i32 := types.NewInt(32)
	i64 := types.NewInt(64)

	printf := mod.NewFunction(types.NewSignature([]types.Type{i8ptr}, types.NewVoid(), true), "printf")
	atol := mod.NewFunction(types.NewSignature([]types.Type{i8ptr}, i64, false), "atol")

	mainFnSig := types.NewSignature([]types.Type{i32, types.NewPointer(i8ptr)}, i32, false)
	mainFn := mod.NewFunction(mainFnSig, "main")
	entry := mainFn.AddBlockAtEnd("entry")
	getMax := mainFn.AddBlockAtEnd("getMax")
	fail := mainFn.AddBlockAtEnd("fail")
	builder := ssa.NewBuilder()

	builder.SetInsertAtBlockStart(entry)
	builder.CreateCondBr(builder.CreateICmp(mainFn.Parameters()[0], ssa.NewIntLiteral(2, i32), ssa.IntSGE, "cmp"), getMax, fail)

	builder.SetInsertAtBlockStart(fail)
	builder.CreateCall(printf, []ssa.Value{builder.CreateConvert(failGlob, i8ptr, ssa.ConvertBitcast, "")}, "")
	builder.CreateRet(ssa.NewIntLiteral(1, i32))

	builder.SetInsertAtBlockStart(getMax)
	secondArg := builder.CreateLoad(builder.CreateGEP(mainFn.Parameters()[1], []ssa.Value{ssa.NewIntLiteral(1, i32)}, ""), "")
	maxorig := builder.CreateCall(atol, []ssa.Value{secondArg}, "max")
	location := builder.CreateAlloc(maxorig.Type(), "")
	builder.CreateStore(location, maxorig)
	max := builder.CreateLoad(location, "")

	mid := mainFn.AddBlockAtEnd("mid")
	exit := mainFn.AddBlockAtEnd("exit")
	builder.CreateBr(mid)
	builder.SetInsertAtBlockStart(mid)

	phi := builder.CreatePhi(i64, "phi")
	phi.AddIncoming(ssa.NewIntLiteral(0, i64), getMax)

	aphi := builder.CreatePhi(i64, "aphi")
	bphi := builder.CreatePhi(i64, "bphi")
	bphi.AddIncoming(ssa.NewIntLiteral(1, i64), getMax)
	aphi.AddIncoming(ssa.NewIntLiteral(0, i64), getMax)
	aphi.AddIncoming(bphi, mid)

	add := builder.CreateBinOp(phi, ssa.NewIntLiteral(1, i64), ssa.BinOpAdd, "add")

	addfib := builder.CreateBinOp(aphi, bphi, ssa.BinOpAdd, "addfib")
	bphi.AddIncoming(addfib, mid)

	builder.CreateCall(printf,
		[]ssa.Value{
			builder.CreateConvert(strGlob, i8ptr, ssa.ConvertBitcast, ""),
			add,
			aphi},
		"")

	phi.AddIncoming(add, mid)
	cmp := builder.CreateICmp(add, max, ssa.IntSGE, "cmp")
	builder.CreateCondBr(cmp, exit, mid)

	builder.SetInsertAtBlockStart(exit)
	builder.CreateRet(builder.CreateConvert(ssa.NewIntLiteral(0, types.NewInt(8)), i32, ssa.ConvertSExt, ""))

	fmt.Println(mod)

	fmt.Println("")

	//analysis.PrintValues(mod, os.Stdout)

	err := validate.Validate(mod)
	if err != nil {
		log.Fatal(err)
	}

	blockCFG := analysis.NewBlockCFG(mainFn)
	err = blockCFG.SaveImage("cfg.svg")
	if err != nil {
		fmt.Println(err)
	}

	domTree := analysis.NewBlockDominatorTree(blockCFG)
	err = domTree.SaveImage("dom.svg")
	if err != nil {
		fmt.Println(err)
	}

	file, err := os.OpenFile("out.s", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	t := amd64.Target{
		Platform: platform.Windows,
	}

	err = t.Generate(file, mod)
	if err != nil {
		log.Fatal(err)
	}
}
