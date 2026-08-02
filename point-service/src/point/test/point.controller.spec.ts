import { HttpException, HttpStatus } from '@nestjs/common';
import { Test, TestingModule } from '@nestjs/testing';
import { CreatePointDto } from '../point.dto';
import { PointController } from '../point.controller';
import { PointService } from '../point.service';

describe('PointController', () => {
  let controller: PointController;

  const mockPointService = {
    getPoint: jest.fn(),
    deductPoint: jest.fn(),
    calculatePoint: jest.fn(),
    calculateDiscount: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [PointController],
      providers: [
        {
          provide: PointService,
          useValue: mockPointService,
        },
      ],
    }).compile();

    controller = module.get<PointController>(PointController);
  });

  it('should be defined', () => {
    expect(controller).toBeDefined();
  });

  it('Create => should create a new point by a given data', async () => {
    // arrange
    const createPointInput = {
      orgId: 1,
      userId: 1,
      amount: 200,
    } as CreatePointDto;

    const createPointResponse = {
      id: 1,
      orgId: 1,
      userId: 1,
      amount: 200,
      created: '2024-08-25T09:06:58',
      updated: '2024-08-25T09:06:58',
    } as CreatePointDto;

    jest
      .spyOn(mockPointService, 'deductPoint')
      .mockReturnValue(createPointResponse);

    // act
    const result = await controller.createPoint(createPointInput);

    // assert
    expect(mockPointService.deductPoint).toBeCalled();
    expect(mockPointService.deductPoint).toBeCalledWith(createPointInput);
    expect(result).toEqual(createPointResponse);
  });

  it('Calculate => should return point calculated by the service from a given amount', () => {
    // arrange
    const amount = '1000';
    const expected = { point: 10 };

    jest.spyOn(mockPointService, 'calculatePoint').mockReturnValue(10);

    // act
    const result = controller.calculatePoint(amount);

    // assert
    expect(mockPointService.calculatePoint).toBeCalled();
    expect(mockPointService.calculatePoint).toBeCalledWith(1000);
    expect(result).toEqual(expected);
  });

  it('Calculate => should throw BAD_REQUEST when amount is missing', () => {
    // arrange
    const amount = undefined;

    // act
    const act = () => controller.calculatePoint(amount);

    // assert
    expect(act).toThrow(HttpException);
    try {
      act();
    } catch (error) {
      expect(error.getStatus()).toEqual(HttpStatus.BAD_REQUEST);
    }
  });

  it('Calculate => should throw BAD_REQUEST when amount is not a number', () => {
    // arrange
    const amount = 'abc';

    // act
    const act = () => controller.calculatePoint(amount);

    // assert
    expect(act).toThrow(HttpException);
    try {
      act();
    } catch (error) {
      expect(error.getStatus()).toEqual(HttpStatus.BAD_REQUEST);
    }
  });

  it('Discount => should return discount calculated by the service from given points and subtotal', () => {
    // arrange
    const points = '160';
    const subtotal = '500';
    const expected = { burn_point: 160, discount: 80 };

    jest
      .spyOn(mockPointService, 'calculateDiscount')
      .mockReturnValue(expected);

    // act
    const result = controller.calculateDiscount(points, subtotal);

    // assert
    expect(mockPointService.calculateDiscount).toBeCalled();
    expect(mockPointService.calculateDiscount).toBeCalledWith(160, 500);
    expect(result).toEqual(expected);
  });

  it('Discount => should throw BAD_REQUEST when points is not a number', () => {
    // arrange
    const points = 'abc';
    const subtotal = '500';

    // act
    const act = () => controller.calculateDiscount(points, subtotal);

    // assert
    expect(act).toThrow(HttpException);
    try {
      act();
    } catch (error) {
      expect(error.getStatus()).toEqual(HttpStatus.BAD_REQUEST);
      expect(error.message).toEqual('points must be a number');
    }
  });

  it('Discount => should throw BAD_REQUEST when subtotal is missing', () => {
    // arrange
    const points = '160';
    const subtotal = undefined;

    // act
    const act = () => controller.calculateDiscount(points, subtotal);

    // assert
    expect(act).toThrow(HttpException);
    try {
      act();
    } catch (error) {
      expect(error.getStatus()).toEqual(HttpStatus.BAD_REQUEST);
      expect(error.message).toEqual('subtotal must be a number');
    }
  });

  it('Find => should return an array of point', async () => {
    //arrange
    const point = {
      id: 2,
      orgId: 1,
      userId: 1,
      amount: 300,
      created: '2024-08-25T09:06:58',
      updated: '2024-08-25T09:06:58',
    };
    const points = [point];
    jest.spyOn(mockPointService, 'getPoint').mockReturnValue(points);

    //act
    const result = await controller.getPoint();

    // assert
    expect(result).toEqual(points);
    expect(mockPointService.getPoint).toBeCalled();
  });
});
