hwaf-rcore2yml
==============

This is a simple WIP converter of `RootCore` requirements files into
(eventually) `YAML` based files.

## Example

```sh
$ hwaf rcore2yml
::: hwaf-rcore2yml
>>> dir="."
::> [PhysicsAnalysis/AnalysisCommon/MissingMassCalculator/cmt/Makefile.RootCore]...
::> [PhysicsAnalysis/AnalysisCommon/PileupReweighting/cmt/Makefile.RootCore]...
req="PhysicsAnalysis/AnalysisCommon/MissingMassCalculator/cmt/Makefile.RootCore"
req="PhysicsAnalysis/AnalysisCommon/MissingMassCalculator/cmt/Makefile.RootCore" [done]
req="PhysicsAnalysis/AnalysisCommon/PileupReweighting/cmt/Makefile.RootCore"
req="PhysicsAnalysis/AnalysisCommon/PileupReweighting/cmt/Makefile.RootCore" [done]
[...]
```
